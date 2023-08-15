package google_gar_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nais/teams-backend/pkg/types"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/reconcilers/google/gar"
	google_workspace_admin_reconciler "github.com/nais/teams-backend/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/pkg/test"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	statusproto "google.golang.org/genproto/googleapis/rpc/status"
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

	deleteCounter int
	delete        func(ctx context.Context, r *artifactregistrypb.DeleteRepositoryRequest) (*longrunningpb.Operation, error)

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

func (f *fakeArtifactRegistry) DeleteRepository(ctx context.Context, r *artifactregistrypb.DeleteRepositoryRequest) (*longrunningpb.Operation, error) {
	f.deleteCounter++
	return f.delete(ctx, r)
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
	if f.delete != nil {
		assert.Equal(t, f.deleteCounter, 1, "mock expected 1 call to delete")
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

	var iamService *iam.Service
	if m.iam != nil {
		var err error
		iamService, err = iam.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(m.iam.URL))
		assert.NoError(t, err)
	}

	return artifactRegistryClient, iamService
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
	groupEmail := "team@example.com"

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
		Description: fmt.Sprintf("Docker repository for team %q. Managed by teams-backend.", team.Slug),
		Labels: map[string]string{
			"team":                         string(team.Slug),
			reconcilers.ManagedByLabelName: reconcilers.ManagedByLabelValue,
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

		err = google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, nil, iamService, log).
			Reconcile(ctx, input)
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
					assert.NotContains(t, req.Policy.Bindings[0].Members, prefix+"/test/archived-repository")
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
						Name: "test/archived-repository",
						Permissions: []*reconcilers.GitHubRepositoryPermission{
							{Name: "push", Granted: true},
							{Name: "admin", Granted: true},
						},
						Archived: true,
					},
					{
						Name: "test/no-permissions-repository",
					},
				}
			}).
			Return(nil).
			Once()
		auditLogger := auditlogger.NewMockAuditLogger(t)

		err = google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, nil, iamService, log).
			Reconcile(ctx, input)
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

		err = google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, artifactregistryClient, iamService, log).
			Reconcile(ctx, input)
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
					assert.Len(t, r.Policy.Bindings, 2)
					assert.Len(t, r.Policy.Bindings[0].Members, 1)
					assert.Len(t, r.Policy.Bindings[1].Members, 1)

					assert.Equal(t, "serviceAccount:"+expectedServiceAccount.Email, r.Policy.Bindings[0].Members[0])
					assert.Equal(t, "roles/artifactregistry.writer", r.Policy.Bindings[0].Role)

					assert.Equal(t, "group:"+groupEmail, r.Policy.Bindings[1].Members[0])
					assert.Equal(t, "roles/artifactregistry.repoAdmin", r.Policy.Bindings[1].Role)

					return &iampb.Policy{}, nil
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
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &groupEmail
			}).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, sqlc.ReconcilerNameGithubTeam, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, google_gar.Name, team.Slug, mock.MatchedBy(func(state reconcilers.GoogleGarState) bool {
				return *state.RepositoryName == garRepositoryParent+"/repositories/"+string(team.Slug)
			})).
			Return(nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)

		err := google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, artifactregistryClient, iamService, log).
			Reconcile(ctx, input)
		assert.NoError(t, err)
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

		err = google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, artifactregistryClient, iamService, log).
			Reconcile(ctx, input)
		assert.ErrorContains(t, err, "abort test")
	})
}

func TestDelete(t *testing.T) {
	const (
		managementProjectID      = "management-project-123"
		workloadIdentityPoolName = "projects/123456789/locations/global/workloadIdentityPools/some-identity-pool"
	)

	ctx := context.Background()
	repositoryName := "some-repo-name-123"
	teamSlug := slug.Slug("my-team")
	correlationID := uuid.New()
	auditLogger := auditlogger.NewMockAuditLogger(t)
	log := logger.NewMockLogger(t)
	mockedClients := mocks{
		artifactRegistry: &fakeArtifactRegistry{},
		iam:              test.HttpServerWithHandlers(t, []http.HandlerFunc{}),
	}
	garClient, iamService := mockedClients.start(t, ctx)

	t.Run("unable to load state", func(t *testing.T) {
		log.
			On("WithComponent", types.ComponentNameGoogleGcpGar).
			Return(log).
			Once()

		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gar.Name, teamSlug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()

		err := google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, garClient, iamService, log).
			Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "load reconciler state for team")
	})

	t.Run("state is missing repository name", func(t *testing.T) {
		log.
			On("WithComponent", types.ComponentNameGoogleGcpGar).
			Return(log).
			Once()

		log.
			On("Warnf",
				"missing repository name in reconciler state for team %q in reconciler %q, assume already deleted",
				teamSlug,
				google_gar.Name).
			Once()

		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gar.Name, teamSlug, mock.Anything).
			Return(nil).
			Once()
		database.On("RemoveReconcilerStateForTeam",
			ctx,
			google_gar.Name,
			teamSlug).
			Return(nil).
			Once()

		err := google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, garClient, iamService, log).
			Delete(ctx, teamSlug, correlationID)
		assert.NoError(t, err)
	})

	t.Run("delete service account fails with unexpected error", func(t *testing.T) {
		log.
			On("WithComponent", types.ComponentNameGoogleGcpGar).
			Return(log).
			Once()

		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gar.Name, teamSlug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGarState)
				state.RepositoryName = &repositoryName
			}).
			Return(nil).
			Once()

		mockedClients := mocks{
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.URL.Path, "management-project-123/serviceAccounts/gar-my-team-a193@management-project-123.iam.gserviceaccount.com")
					w.WriteHeader(http.StatusInternalServerError)
				},
			}),
		}
		garClient, iamService := mockedClients.start(t, ctx)

		err := google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, garClient, iamService, log).
			Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "delete service account")
	})

	t.Run("service account does not exist, and delete repo request fails", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gar.Name, teamSlug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGarState)
				state.RepositoryName = &repositoryName
			}).
			Return(nil).
			Once()

		mockedClients := mocks{
			artifactRegistry: &fakeArtifactRegistry{
				delete: func(ctx context.Context, req *artifactregistrypb.DeleteRepositoryRequest) (*longrunningpb.Operation, error) {
					return nil, fmt.Errorf("some error")
				},
			},
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				},
			}),
		}
		garClient, iamService := mockedClients.start(t, ctx)

		testLogger, logs := logrustest.NewNullLogger()

		log := logger.NewMockLogger(t)
		log.
			On("WithComponent", types.ComponentNameGoogleGcpGar).
			Return(log).
			Once()
		log.
			On("WithTeamSlug", string(teamSlug)).
			Return(log).
			Once()
		log.
			On("WithError", mock.Anything).
			Return(&logrus.Entry{Logger: testLogger}).
			Once()

		err := google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, garClient, iamService, log).
			Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "delete GAR repository for team")
		assert.Contains(t, logs.Entries[0].Message, "does not exist")
	})

	t.Run("delete repo operation fails", func(t *testing.T) {
		log.
			On("WithComponent", types.ComponentNameGoogleGcpGar).
			Return(log).
			Once()

		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gar.Name, teamSlug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGarState)
				state.RepositoryName = &repositoryName
			}).
			Return(nil).
			Once()

		mockedClients := mocks{
			artifactRegistry: &fakeArtifactRegistry{
				delete: func(ctx context.Context, req *artifactregistrypb.DeleteRepositoryRequest) (*longrunningpb.Operation, error) {
					return &longrunningpb.Operation{
						Done: true,
						Result: &longrunningpb.Operation_Error{
							Error: &statusproto.Status{
								Code:    int32(codes.NotFound),
								Message: "not found",
							},
						},
					}, nil
				},
			},
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				},
			}),
		}
		garClient, iamService := mockedClients.start(t, ctx)

		err := google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, garClient, iamService, log).
			Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "wait for GAR repository deletion")
	})

	t.Run("successful delete", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gar.Name, teamSlug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleGarState)
				state.RepositoryName = &repositoryName
			}).
			Return(nil).
			Once()
		database.
			On("RemoveReconcilerStateForTeam", ctx, google_gar.Name, teamSlug).
			Return(nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.EXPECT().
			Logf(
				ctx,
				mock.MatchedBy(func(targets []auditlogger.Target) bool {
					return targets[0].Identifier == string(teamSlug)
				}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
					return fields.Action == types.AuditActionGoogleGarDelete && fields.CorrelationID == correlationID
				}),
				mock.MatchedBy(func(msg string) bool {
					return strings.HasPrefix(msg, "Delete GAR repository")
				}),
				repositoryName,
			).
			Return().
			Once()

		log.
			On("WithComponent", types.ComponentNameGoogleGcpGar).
			Return(log).
			Once()

		mockedClients := mocks{
			artifactRegistry: &fakeArtifactRegistry{
				delete: func(ctx context.Context, req *artifactregistrypb.DeleteRepositoryRequest) (*longrunningpb.Operation, error) {
					assert.Equal(t, repositoryName, req.Name)
					return &longrunningpb.Operation{
						Done:   true,
						Result: &longrunningpb.Operation_Response{},
					}, nil
				},
			},
			iam: test.HttpServerWithHandlers(t, []http.HandlerFunc{
				func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				},
			}),
		}
		garClient, iamService := mockedClients.start(t, ctx)

		err := google_gar.
			New(auditLogger, database, managementProjectID, workloadIdentityPoolName, garClient, iamService, log).
			Delete(ctx, teamSlug, correlationID)
		assert.NoError(t, err)
	})
}
