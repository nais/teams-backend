package gar_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/google/gar"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type fakeArtifaceRegistry struct {
	createCounter int
	create        func(ctx context.Context, r *artifactregistrypb.CreateRepositoryRequest) (*longrunningpb.Operation, error)

	getCounter int
	get        func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error)

	updateCounter int
	update        func(ctx context.Context, r *artifactregistrypb.UpdateRepositoryRequest) (*artifactregistrypb.Repository, error)

	artifactregistrypb.UnimplementedArtifactRegistryServer
}

type mocks struct {
	artifaceRegistry *fakeArtifaceRegistry
	iam              *httptest.Server
}

func (f *fakeArtifaceRegistry) CreateRepository(ctx context.Context, r *artifactregistrypb.CreateRepositoryRequest) (*longrunningpb.Operation, error) {
	f.createCounter++
	return f.create(ctx, r)
}

func (f *fakeArtifaceRegistry) GetRepository(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
	f.getCounter++
	return f.get(ctx, r)
}

func (f *fakeArtifaceRegistry) UpdateRepository(ctx context.Context, r *artifactregistrypb.UpdateRepositoryRequest) (*artifactregistrypb.Repository, error) {
	f.updateCounter++
	return f.update(ctx, r)
}

func (f *fakeArtifaceRegistry) assert(t *testing.T) {
	if f.create != nil {
		assert.Equal(t, f.createCounter, 1, "expected 1 call to create")
	}
	if f.update != nil {
		assert.Equal(t, f.updateCounter, 1, "expected 1 call to update")
	}
	if f.get != nil {
		assert.Equal(t, f.getCounter, 1, "expected 1 call to get")
	}
}

func (m *mocks) start(t *testing.T, ctx context.Context) (*artifactregistry.Client, *iam.Service) {
	t.Helper()

	var artifactRegistryClient *artifactregistry.Client
	if m.artifaceRegistry != nil {
		l, err := net.Listen("tcp", "localhost:0")
		assert.NoError(t, err)

		srv := grpc.NewServer()
		artifactregistrypb.RegisterArtifactRegistryServer(srv, m.artifaceRegistry)
		go func() {
			if err := srv.Serve(l); err != nil {
				panic(err)
			}
		}()
		t.Cleanup(func() {
			m.artifaceRegistry.assert(t)
			srv.Stop()
		})

		artifactRegistryClient, err = artifactregistry.NewClient(ctx,
			option.WithEndpoint(l.Addr().String()),
			option.WithoutAuthentication(),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		)
		assert.NoError(t, err)
	}

	fakeIamService, err := iam.NewService(ctx, option.WithEndpoint(m.iam.URL))
	assert.NoError(t, err)

	return artifactRegistryClient, fakeIamService
}

// fakeIamServer := test.HttpServerWithHandlers(t, []http.HandlerFunc{
//   func(w http.ResponseWriter, r *http.Request) {},
// })

func TestReconcile(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)

	const (
		managementProjectID = "management-project-123"
		abortReconcilerCode = 418
	)

	correlationID := uuid.New()
	team := db.Team{Team: &sqlc.Team{Slug: slug.Slug("team")}}
	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}

	ctx := context.Background()
	// repositoryName := fmt.Sprintf("projects/%s/locations/europe-north1/repositories/%s", managementProjectID, string(slug))
	// repositoryDescription := fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", string(slug))

	t.Run("when service account does not exist, create it", func(t *testing.T) {
		mocks := mocks{
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(404)
				},
				func(w http.ResponseWriter, r *http.Request) {
					expected := &iam.ServiceAccount{
						Description: fmt.Sprintf("Service Account used to push images to Google Artifact Registry for %s", team.Slug),
						DisplayName: fmt.Sprintf("Artifact Pusher for %s", team.Slug),
					}

					var req iam.CreateServiceAccountRequest
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
					assert.Equal(t, expected.Description, req.ServiceAccount.Description)
					assert.Equal(t, expected.DisplayName, req.ServiceAccount.DisplayName)
					w.WriteHeader(abortReconcilerCode) // abort test - we have asserted what we are interested in already
				},
			}),
		}
		_, iamService := mocks.start(t, ctx)
		database := db.NewMockDatabase(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)

		reconciler := gar.New(auditLogger, database, managementProjectID, nil, iamService, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, fmt.Sprintf("googleapi: got HTTP response code %d", abortReconcilerCode))
	})

	t.Run("after getOrCreateServiceAccount, set policy", func(t *testing.T) {
		email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", team.Slug, managementProjectID)
		expected := &iam.ServiceAccount{
			Email:       email,
			Name:        fmt.Sprintf("projects/%s/serviceAccounts/%s", managementProjectID, email),
			Description: fmt.Sprintf("Service Account used to push images to Google Artifact Registry for %s", team.Slug),
			DisplayName: fmt.Sprintf("Artifact Pusher for %s", team.Slug),
		}

		mocks := mocks{
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, json.NewEncoder(w).Encode(&expected))
				},
				func(w http.ResponseWriter, r *http.Request) {
					var req iam.SetIamPolicyRequest
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
					assert.Contains(t, r.URL.Path, expected.Name)
					w.WriteHeader(abortReconcilerCode)
				},
			}),
		}
		_, iamService := mocks.start(t, ctx)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, sqlc.ReconcilerNameGithubTeam, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GitHubState)
				state.Repositories = []*reconcilers.GitHubRepository{
					{
						Name: "test/repository",
						Permissions: []*reconcilers.GitHubRepositoryPermission{
							{Name: "push", Granted: true},
						},
					},
				}
			}).
			Return(nil).
			Once()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		reconciler := gar.New(auditLogger, database, managementProjectID, nil, iamService, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, fmt.Sprintf("googleapi: got HTTP response code %d", abortReconcilerCode))
	})

	t.Run("after getOrCreateServiceAccount, set iam ", func(t *testing.T) {
		email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", team.Slug, managementProjectID)
		expected := &iam.ServiceAccount{
			Email:       email,
			Name:        fmt.Sprintf("projects/%s/serviceAccounts/%s", managementProjectID, email),
			Description: fmt.Sprintf("Service Account used to push images to Google Artifact Registry for %s", team.Slug),
			DisplayName: fmt.Sprintf("Artifact Pusher for %s", team.Slug),
		}
		mocks := mocks{
			artifaceRegistry: &fakeArtifaceRegistry{
				get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
					return nil, fmt.Errorf("test error")
				},
			},
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, json.NewEncoder(w).Encode(expected))
				},
				func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode(&iam.Policy{})
				},
			}),
		}
		artifactregistryClient, iamService := mocks.start(t, ctx)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, sqlc.ReconcilerNameGithubTeam, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GitHubState)
				state.Repositories = []*reconcilers.GitHubRepository{
					{
						Name: "test/repository",
						Permissions: []*reconcilers.GitHubRepositoryPermission{
							{
								Name:    "push",
								Granted: true,
							},
						},
					},
				}
			}).
			Return(nil).
			Once()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		reconciler := gar.New(auditLogger, database, managementProjectID, artifactregistryClient, iamService, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "test error")
	})

	//		t.Run("create when no repository already exists", func(t *testing.T) {
	//
	//			fake := &fakeArtifaceRegistry{
	//				get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
	//					return nil, status.Errorf(codes.NotFound, "test error")
	//				},
	//				create: func(ctx context.Context, r *artifactregistrypb.CreateRepositoryRequest) (*longrunningpb.Operation, error) {
	//					assert.Equal(t, repositoryName, r.Repository.Name)
	//					assert.Equal(t, repositoryDescription, r.Repository.Description)
	//					assert.Equal(t, artifactregistrypb.Repository_DOCKER, r.Repository.Format)
	//					assert.Equal(t, string(slug), r.Repository.Labels["team"])
	//
	//					payload := anypb.Any{}
	//					err := anypb.MarshalFrom(&payload, r.Repository, proto.MarshalOptions{})
	//					assert.NoError(t, err)
	//
	//					return &longrunningpb.Operation{
	//						Done: true,
	//						Result: &longrunningpb.Operation_Response{
	//							Response: &payload,
	//						},
	//					}, nil
	//				},
	//			}
	//
	//			client, err := fake.start(t, ctx)
	//			assert.NoError(t, err)
	//			defer fake.stop()
	//
	//			auditLogger := auditlogger.NewMockAuditLogger(t)
	//			reconciler := gar.New(auditLogger, managementProjectID, client, log)
	//			err = reconciler.Reconcile(ctx, input)
	//			assert.NoError(t, err)
	//			fake.assert(t)
	//		})
	//
	//		t.Run("no update when existing repository is up to date", func(t *testing.T) {
	//			repo := &artifactregistrypb.Repository{
	//				Name:        repositoryName,
	//				Format:      artifactregistrypb.Repository_DOCKER,
	//				Description: repositoryDescription,
	//				Labels:      map[string]string{"team": string(slug)},
	//			}
	//			fake := &fakeArtifaceRegistry{
	//				get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
	//					return repo, nil
	//				},
	//			}
	//
	//			client, err := fake.start(t, ctx)
	//			assert.NoError(t, err)
	//			defer fake.stop()
	//
	//			auditLogger := auditlogger.NewMockAuditLogger(t)
	//			reconciler := gar.New(auditLogger, managementProjectID, client, log)
	//			err = reconciler.Reconcile(ctx, input)
	//			assert.NoError(t, err)
	//			fake.assert(t)
	//		})
	//
	//		t.Run("update when existing repository has incorrect desc", func(t *testing.T) {
	//			repo := &artifactregistrypb.Repository{
	//				Name:        repositoryName,
	//				Format:      artifactregistrypb.Repository_DOCKER,
	//				Description: repositoryDescription + "invalid",
	//				Labels:      map[string]string{"team": string(slug)},
	//			}
	//
	//			fake := &fakeArtifaceRegistry{
	//				get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
	//					return repo, nil
	//				},
	//				update: func(ctx context.Context, r *artifactregistrypb.UpdateRepositoryRequest) (*artifactregistrypb.Repository, error) {
	//					assert.Len(t, r.UpdateMask.Paths, 1)
	//					r.Repository.Description = repositoryDescription
	//
	//					return r.Repository, nil
	//				},
	//			}
	//
	//			client, err := fake.start(t, ctx)
	//			assert.NoError(t, err)
	//			defer fake.stop()
	//
	//			auditLogger := auditlogger.NewMockAuditLogger(t)
	//			reconciler := gar.New(auditLogger, managementProjectID, client, log)
	//			err = reconciler.Reconcile(ctx, input)
	//			assert.NoError(t, err)
	//			fake.assert(t)
	//	})
}
