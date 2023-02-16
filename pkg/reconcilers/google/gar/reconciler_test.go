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
	"cloud.google.com/go/iam/apiv1/iampb"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type fakeArtifactRegistry struct {
	createCounter int
	create        func(ctx context.Context, r *artifactregistrypb.CreateRepositoryRequest) (*longrunningpb.Operation, error)

	getCounter int
	get        func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error)

	updateCounter int
	update        func(ctx context.Context, r *artifactregistrypb.UpdateRepositoryRequest) (*artifactregistrypb.Repository, error)

	setIamPolicy        func(context.Context, *iampb.SetIamPolicyRequest) (*iampb.Policy, error)
	setIamPolicyCounter int

	artifactregistrypb.UnimplementedArtifactRegistryServer
}

type mocks struct {
	artifactRegistry *fakeArtifactRegistry
	iam              *httptest.Server
}

func (f *fakeArtifactRegistry) CreateRepository(ctx context.Context, r *artifactregistrypb.CreateRepositoryRequest) (*longrunningpb.Operation, error) {
	f.createCounter++
	return f.create(ctx, r)
}

func (f *fakeArtifactRegistry) GetRepository(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
	f.getCounter++
	return f.get(ctx, r)
}

func (f *fakeArtifactRegistry) UpdateRepository(ctx context.Context, r *artifactregistrypb.UpdateRepositoryRequest) (*artifactregistrypb.Repository, error) {
	f.updateCounter++
	return f.update(ctx, r)
}

func (f *fakeArtifactRegistry) SetIamPolicy(ctx context.Context, r *iampb.SetIamPolicyRequest) (*iampb.Policy, error) {
	f.setIamPolicyCounter++
	return f.setIamPolicy(ctx, r)
}

func (f *fakeArtifactRegistry) assert(t *testing.T) {
	if f.create != nil {
		assert.Equal(t, f.createCounter, 1, "mock expected 1 call to create")
	}
	if f.update != nil {
		assert.Equal(t, f.updateCounter, 1, "mock expected 1 call to update")
	}
	if f.get != nil {
		assert.Equal(t, f.getCounter, 1, "mock expected 1 call to get")
	}
	if f.setIamPolicy != nil {
		assert.Equal(t, f.setIamPolicyCounter, 1, "mock expected 1 call to setIamPolicy")
	}
}

func (m *mocks) start(t *testing.T, ctx context.Context) (*artifactregistry.Client, *iam.Service) {
	t.Helper()

	var artifactRegistryClient *artifactregistry.Client
	if m.artifactRegistry != nil {
		l, err := net.Listen("tcp", "localhost:0")
		assert.NoError(t, err)

		srv := grpc.NewServer()
		artifactregistrypb.RegisterArtifactRegistryServer(srv, m.artifactRegistry)
		go func() {
			if err := srv.Serve(l); err != nil {
				panic(err)
			}
		}()
		t.Cleanup(func() {
			m.artifactRegistry.assert(t)
			srv.Stop()
		})

		artifactRegistryClient, err = artifactregistry.NewClient(ctx,
			option.WithEndpoint(l.Addr().String()),
			option.WithoutAuthentication(),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		)
		assert.NoError(t, err)
	}

	fakeIamService, err := iam.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(m.iam.URL))
	assert.NoError(t, err)

	return artifactRegistryClient, fakeIamService
}

func TestReconcile(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)

	const (
		managementProjectID      = "management-project-123"
		workloadIdentityPoolName = "projects/123456789/locations/global/workloadIdentityPools/some-identity-pool"
		abortReconcilerCode      = 418
	)

	abortTestErr := fmt.Errorf("abort test")

	correlationID := uuid.New()
	team := db.Team{Team: &sqlc.Team{Slug: slug.Slug("team")}}
	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}

	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", team.Slug, managementProjectID)
	expectedServiceAccount := &iam.ServiceAccount{
		Email:       email,
		Name:        fmt.Sprintf("projects/%s/serviceAccounts/%s", managementProjectID, email),
		Description: fmt.Sprintf("Service Account used to push images to Google Artifact Registry for %s", team.Slug),
		DisplayName: fmt.Sprintf("Artifact Pusher for %s", team.Slug),
	}

	garRepositoryParent := fmt.Sprintf("projects/%s/locations/europe-north1", managementProjectID)
	expectedRepository := artifactregistrypb.Repository{
		Name:        fmt.Sprintf("%s/repositories/%s", garRepositoryParent, string(team.Slug)),
		Format:      artifactregistrypb.Repository_DOCKER,
		Description: fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", team.Slug),
		Labels: map[string]string{
			"team": string(team.Slug),
		},
	}

	ctx := context.Background()

	t.Run("when service account does not exist, create it", func(t *testing.T) {
		mocks := mocks{
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(404)
				},
				func(w http.ResponseWriter, r *http.Request) {
					var req iam.CreateServiceAccountRequest
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
					assert.Equal(t, expectedServiceAccount.Description, req.ServiceAccount.Description)
					assert.Equal(t, expectedServiceAccount.DisplayName, req.ServiceAccount.DisplayName)
					w.WriteHeader(abortReconcilerCode) // abort test - we have asserted what we are interested in already
				},
			}),
		}
		_, iamService := mocks.start(t, ctx)
		database := db.NewMockDatabase(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)

		reconciler := gar.New(auditLogger, database, managementProjectID, workloadIdentityPoolName, nil, iamService, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, fmt.Sprintf("googleapi: got HTTP response code %d", abortReconcilerCode))
	})

	t.Run("after getOrCreateServiceAccount, set policy", func(t *testing.T) {
		mocks := mocks{
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, json.NewEncoder(w).Encode(&expectedServiceAccount))
				},
				func(w http.ResponseWriter, r *http.Request) {
					var req iam.SetIamPolicyRequest
					prefix := "principalSet://iam.googleapis.com/" + workloadIdentityPoolName + "/attribute.repository"
					assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
					assert.Contains(t, r.URL.Path, expectedServiceAccount.Name)
					assert.Contains(t, req.Policy.Bindings[0].Members, prefix+"/test/repository")
					assert.Contains(t, req.Policy.Bindings[0].Members, prefix+"/test/admin-repository")
					assert.NotContains(t, req.Policy.Bindings[0].Members, prefix+"/test/ro-repository")
					assert.NotContains(t, req.Policy.Bindings[0].Members, prefix+"/test/no-permissions-repository")
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
					{
						Name: "test/ro-repository",
						Permissions: []*reconcilers.GitHubRepositoryPermission{
							{Name: "push", Granted: false},
						},
					},
					{
						Name: "test/admin-repository",
						Permissions: []*reconcilers.GitHubRepositoryPermission{
							{Name: "push", Granted: true},
							{Name: "admin", Granted: true},
						},
					},
					{
						Name: "test/no-permissions-repository",
					},
				}
			}).
			Return(nil).
			Once()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		reconciler := gar.New(auditLogger, database, managementProjectID, workloadIdentityPoolName, nil, iamService, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, fmt.Sprintf("googleapi: got HTTP response code %d", abortReconcilerCode))
	})

	t.Run("if no gar repository exists, create it", func(t *testing.T) {
		mocks := mocks{
			artifactRegistry: &fakeArtifactRegistry{
				get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
					return nil, status.Error(codes.NotFound, "not found")
				},
				create: func(ctx context.Context, r *artifactregistrypb.CreateRepositoryRequest) (*longrunningpb.Operation, error) {
					assert.Equal(t, r.Repository.Name, expectedRepository.Name)
					assert.Equal(t, r.Repository.Description, expectedRepository.Description)
					assert.Equal(t, r.Parent, garRepositoryParent)
					assert.Equal(t, r.Repository.Format, expectedRepository.Format)

					payload := anypb.Any{}
					err := anypb.MarshalFrom(&payload, r.Repository, proto.MarshalOptions{})
					assert.NoError(t, err)

					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Response{
							Response: &payload,
						},
					}, abortTestErr
				},
			},
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				// get service account
				func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, json.NewEncoder(w).Encode(expectedServiceAccount))
				},
				// set iam policy
				func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, json.NewEncoder(w).Encode(&iam.Policy{}))
				},
			}),
		}
		artifactregistryClient, iamService := mocks.start(t, ctx)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, sqlc.ReconcilerNameGithubTeam, team.Slug, mock.Anything).
			Return(nil).
			Once()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		reconciler := gar.New(auditLogger, database, managementProjectID, workloadIdentityPoolName, artifactregistryClient, iamService, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "abort test")
	})

	t.Run("if gar repository exists, set iam policy", func(t *testing.T) {
		mocks := mocks{
			artifactRegistry: &fakeArtifactRegistry{
				get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
					return &expectedRepository, nil
				},
				setIamPolicy: func(ctx context.Context, r *iampb.SetIamPolicyRequest) (*iampb.Policy, error) {
					assert.Equal(t, expectedRepository.Name, r.Resource)
					assert.Equal(t, "serviceAccount:"+expectedServiceAccount.Email, r.Policy.Bindings[0].Members[0])
					assert.Equal(t, "roles/artifactregistry.writer", r.Policy.Bindings[0].Role)

					return &iampb.Policy{}, abortTestErr
				},
			},
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				// get service account
				func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, json.NewEncoder(w).Encode(expectedServiceAccount))
				},
				// set iam policy
				func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, json.NewEncoder(w).Encode(&iam.Policy{}))
				},
			}),
		}

		artifactregistryClient, iamService := mocks.start(t, ctx)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, sqlc.ReconcilerNameGithubTeam, team.Slug, mock.Anything).
			Return(nil).
			Once()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		reconciler := gar.New(auditLogger, database, managementProjectID, workloadIdentityPoolName, artifactregistryClient, iamService, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "abort test")
	})

	t.Run("gar repository exists, but has outdated info", func(t *testing.T) {
		mocks := mocks{
			artifactRegistry: &fakeArtifactRegistry{
				get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
					assert.Equal(t, expectedRepository.Name, r.Name)

					repo := expectedRepository
					repo.Description = "some incorrect description"
					repo.Labels = map[string]string{
						"team": "some-incorrect-team",
					}
					return &repo, nil
				},
				update: func(ctx context.Context, r *artifactregistrypb.UpdateRepositoryRequest) (*artifactregistrypb.Repository, error) {
					assert.Equal(t, expectedRepository.Description, r.Repository.Description)
					assert.Equal(t, expectedRepository.Name, r.Repository.Name)
					assert.Equal(t, string(team.Slug), r.Repository.Labels["team"])

					return nil, abortTestErr
				},
			},
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				// get service account
				func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, json.NewEncoder(w).Encode(expectedServiceAccount))
				},
				// set iam policy
				func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(t, json.NewEncoder(w).Encode(&iam.Policy{}))
				},
			}),
		}

		artifactregistryClient, iamService := mocks.start(t, ctx)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, sqlc.ReconcilerNameGithubTeam, team.Slug, mock.Anything).
			Return(nil).
			Once()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		reconciler := gar.New(auditLogger, database, managementProjectID, workloadIdentityPoolName, artifactregistryClient, iamService, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "abort test")
	})
}
