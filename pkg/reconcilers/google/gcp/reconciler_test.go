package google_gcp_reconciler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/gcp"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/reconcilers/google/gcp"
	"github.com/nais/teams-backend/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
)

const (
	env              = "prod"
	teamFolderID     = 123
	clusterProjectID = "some-project-123"
	tenantName       = "example"
	tenantDomain     = "example.com"
	cnrmRoleName     = "organizations/123/roles/name"
	billingAccount   = "billingAccounts/123"
	numberOfAPIs     = 12
)

var (
	clusters = gcp.Clusters{
		env: {
			TeamsFolderID: teamFolderID,
			ProjectID:     clusterProjectID,
		},
	}
	teamSlug      = slug.Slug("slug")
	correlationID = uuid.New()
	team          = db.Team{Team: &sqlc.Team{Slug: teamSlug}}
	input         = reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}
)

func TestReconcile(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)

	t.Run("fail early when unable to load reconciler state", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithComponentName", types.ComponentNameGoogleGcpProject).
			Return(auditLogger).
			Once()

		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}

		err := google_gcp_reconciler.
			New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, log, nil).
			Reconcile(ctx, input)
		assert.ErrorContains(t, err, "load system state")
	})

	t.Run("fail early when unable to load google workspace state", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithComponentName", types.ComponentNameGoogleGcpProject).
			Return(auditLogger).
			Once()
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}

		err := google_gcp_reconciler.
			New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, log, nil).
			Reconcile(ctx, input)
		assert.ErrorContains(t, err, "load system state")
	})

	t.Run("fail early when google workspace state is missing group email", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithComponentName", types.ComponentNameGoogleGcpProject).
			Return(auditLogger).
			Once()
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}

		err := google_gcp_reconciler.
			New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, log, nil).
			Reconcile(ctx, input)
		assert.ErrorContains(t, err, "no Google Workspace group exists")
	})

	t.Run("no error when we have no clusters", func(t *testing.T) {
		ctx := context.Background()
		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithComponentName", types.ComponentNameGoogleGcpProject).
			Return(auditLogger).
			Once()
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				email := "mail@example.com"
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &email
			}).
			Return(nil).
			Once()
		gcpServices := &google_gcp_reconciler.GcpServices{}

		err := google_gcp_reconciler.
			New(database, auditLogger, gcp.Clusters{}, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, log, nil).
			Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("full reconcile, no existing project state", func(t *testing.T) {
		clusters := gcp.Clusters{
			env: gcp.Cluster{
				TeamsFolderID: teamFolderID,
				ProjectID:     clusterProjectID,
			},
		}
		const expectedTeamProjectID = "slug-prod-ea99"
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, team.Slug, mock.Anything).
			Run(func(args mock.Arguments) {
				email := "mail@example.com"
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &email
			}).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, team.Slug, mock.MatchedBy(func(state *reconcilers.GoogleGcpProjectState) bool {
				return state.Projects[env].ProjectID == expectedTeamProjectID
			})).
			Return(nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On("WithComponentName", types.ComponentNameGoogleGcpProject).
			Return(auditLogger).
			Once()
		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == correlationID && fields.Action == types.AuditActionGoogleGcpProjectCreateProject
			}), mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "Created GCP project")
			}), expectedTeamProjectID, teamSlug, env).
			Return(nil).
			Once()
		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == correlationID && fields.Action == types.AuditActionGoogleGcpProjectSetBillingInfo
			}), mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "Set billing info")
			}), expectedTeamProjectID).
			Return(nil).
			Once()
		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == correlationID && fields.Action == types.AuditActionGoogleGcpProjectCreateCnrmServiceAccount
			}), mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "Created CNRM service account")
			}), teamSlug, expectedTeamProjectID).
			Return(nil).
			Once()
		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == correlationID && fields.Action == types.AuditActionGoogleGcpProjectAssignPermissions
			}), mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "Assigned GCP project IAM permissions")
			}), expectedTeamProjectID).
			Return(nil).
			Once()
		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == correlationID && fields.Action == types.AuditActionGoogleGcpProjectEnableGoogleApis
			}), mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "Enable Google API")
			}), mock.AnythingOfType("string"), expectedTeamProjectID).
			Return(nil).
			Times(numberOfAPIs)

		srv := test.HttpServerWithHandlers(t, []http.HandlerFunc{
			// create project request
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				payload := cloudresourcemanager.Project{}
				json.NewDecoder(r.Body).Decode(&payload)
				assert.Equal(t, "slug-prod", payload.DisplayName)
				assert.Equal(t, "folders/123", payload.Parent)
				assert.Equal(t, expectedTeamProjectID, payload.ProjectId)

				project := cloudresourcemanager.Project{
					Name:      payload.DisplayName,
					ProjectId: payload.ProjectId,
				}
				projectJson, _ := project.MarshalJSON()

				op := cloudresourcemanager.Operation{
					Done:     true,
					Response: projectJson,
				}
				resp, _ := op.MarshalJSON()
				w.Write(resp)
			},

			// set project labels
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPatch, r.Method)
				payload := cloudresourcemanager.Project{}
				json.NewDecoder(r.Body).Decode(&payload)
				assert.Equal(t, env, payload.Labels["environment"])
				assert.Equal(t, string(teamSlug), payload.Labels["team"])
				assert.Equal(t, tenantName, payload.Labels["tenant"])
				assert.Equal(t, reconcilers.ManagedByLabelValue, payload.Labels[reconcilers.ManagedByLabelName])

				project, _ := payload.MarshalJSON()
				op := cloudresourcemanager.Operation{
					Done:     true,
					Response: project,
				}
				resp, _ := op.MarshalJSON()
				w.Write(resp)
			},

			// get existing billing info
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				info := cloudbilling.ProjectBillingInfo{}
				resp, _ := info.MarshalJSON()
				w.Write(resp)
			},

			// update billing info
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				payload := cloudbilling.ProjectBillingInfo{}
				json.NewDecoder(r.Body).Decode(&payload)
				assert.Equal(t, billingAccount, payload.BillingAccountName)

				info := cloudbilling.ProjectBillingInfo{}
				resp, _ := info.MarshalJSON()
				w.Write(resp)
			},

			// get existing CNRM service account
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(404)
			},

			// create CNRM service account
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				payload := iam.CreateServiceAccountRequest{}
				json.NewDecoder(r.Body).Decode(&payload)
				assert.Equal(t, reconcilers.CnrmServiceAccountAccountID, payload.AccountId)
				assert.Equal(t, "CNRM service account", payload.ServiceAccount.DisplayName)

				sa := iam.ServiceAccount{
					Name:  "projects/some-project-123/serviceAccounts/cnrm@some-project-123.iam.gserviceaccount.com",
					Email: "cnrm@some-project-123.iam.gserviceaccount.com",
				}
				resp, _ := sa.MarshalJSON()
				w.Write(resp)
			},

			// set workload identity for service account
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				payload := iam.SetIamPolicyRequest{}
				json.NewDecoder(r.Body).Decode(&payload)
				assert.Equal(t, "serviceAccount:some-project-123.svc.id.goog[cnrm-system/cnrm-controller-manager-slug]", payload.Policy.Bindings[0].Members[0])
				assert.Equal(t, "roles/iam.workloadIdentityUser", payload.Policy.Bindings[0].Role)

				policy := iam.Policy{}
				resp, _ := policy.MarshalJSON()
				w.Write(resp)
			},

			// get existing IAM policy for the team project
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				policy := iam.Policy{}
				resp, _ := policy.MarshalJSON()
				w.Write(resp)
			},

			// set updated IAM policy for the team project
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				payload := iam.SetIamPolicyRequest{}
				json.NewDecoder(r.Body).Decode(&payload)
				expectedBindings := map[string]string{
					payload.Policy.Bindings[0].Role: payload.Policy.Bindings[0].Members[0],
					payload.Policy.Bindings[1].Role: payload.Policy.Bindings[1].Members[0],
				}
				assert.Equal(t, "group:mail@example.com", expectedBindings["roles/owner"])
				assert.Equal(t, "serviceAccount:cnrm@some-project-123.iam.gserviceaccount.com", expectedBindings[cnrmRoleName])

				policy := iam.Policy{}
				resp, _ := policy.MarshalJSON()
				w.Write(resp)
			},

			// list existing Google APIs for the team project
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				services := serviceusage.ListServicesResponse{}
				resp, _ := services.MarshalJSON()
				w.Write(resp)
			},

			// enable Google APIs for the team project
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				payload := serviceusage.BatchEnableServicesRequest{}
				json.NewDecoder(r.Body).Decode(&payload)
				assert.Len(t, payload.ServiceIds, numberOfAPIs)

				op := serviceusage.Operation{Done: true}
				resp, _ := op.MarshalJSON()
				w.Write(resp)
			},

			// list firewall rules for project
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/projects/"+expectedTeamProjectID+"/global/firewalls", r.URL.Path)

				list := compute.FirewallList{
					Items: []*compute.Firewall{
						{
							Name:     "default-allow-ssh",
							Priority: 65534,
						},
					},
				}

				resp, _ := list.MarshalJSON()
				w.Write(resp)
			},

			// delete default firewall rule
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Equal(t, "/projects/"+expectedTeamProjectID+"/global/firewalls/default-allow-ssh", r.URL.Path)

				op := compute.Operation{Name: "operation-name", Status: "RUNNING"}
				resp, _ := op.MarshalJSON()
				w.Write(resp)
			},

			// wait for operation to complete
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/projects/"+expectedTeamProjectID+"/global/operations/operation-name/wait", r.URL.Path)

				op := compute.Operation{Name: "operation-name", Status: "DONE"}
				resp, _ := op.MarshalJSON()
				w.Write(resp)
			},
		})
		defer srv.Close()

		cloudBillingService, _ := cloudbilling.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
		cloudResourceManagerService, _ := cloudresourcemanager.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
		iamService, _ := iam.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
		serviceUsageService, _ := serviceusage.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
		computeService, _ := compute.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(srv.URL))

		gcpServices := &google_gcp_reconciler.GcpServices{
			CloudBillingProjectsService:           cloudBillingService.Projects,
			CloudResourceManagerProjectsService:   cloudResourceManagerService.Projects,
			CloudResourceManagerOperationsService: cloudResourceManagerService.Operations,
			IamProjectsServiceAccountsService:     iamService.Projects.ServiceAccounts,
			ServiceUsageService:                   serviceUsageService.Services,
			ServiceUsageOperationsService:         serviceUsageService.Operations,
			FirewallService:                       computeService.Firewalls,
			ComputeGlobalOperationsService:        computeService.GlobalOperations,
		}

		err = google_gcp_reconciler.
			New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, log, nil).
			Reconcile(ctx, input)
		assert.NoError(t, err)
	})
}

func TestGenerateProjectID(t *testing.T) {
	// different organization names don't show up in name, but are reflected in the hash
	assert.Equal(t, "happyteam-prod-488a", google_gcp_reconciler.GenerateProjectID("nais.io", "production", "happyteam"))
	assert.Equal(t, "happyteam-prod-5534", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "happyteam"))

	// environments that get truncated produce different hashes
	assert.Equal(t, "sadteam-prod-04d4", google_gcp_reconciler.GenerateProjectID("nais.io", "production", "sadteam"))
	assert.Equal(t, "sadteam-prod-6ce6", google_gcp_reconciler.GenerateProjectID("nais.io", "producers", "sadteam"))

	// team names that get truncated produce different hashes
	assert.Equal(t, "happyteam-is-very-ha-prod-4b2d", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "happyteam-is-very-happy"))
	assert.Equal(t, "happyteam-is-very-ha-prod-4801", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "happyteam-is-very-happy-and-altogether-too-long"))

	// project id with double hyphens
	assert.Equal(t, "hapyteam-is-very-ha-prod-fd5d", google_gcp_reconciler.GenerateProjectID("bais.io", "production", "hapyteam-is-very-ha-a"))

	// environment with hyphen as 4th character in environment
	assert.Equal(t, "hapyteam-is-happy-pro-2a15", google_gcp_reconciler.GenerateProjectID("bais.io", "pro-duction", "hapyteam-is-happy"))
}

func TestGetProjectDisplayName(t *testing.T) {
	tests := []struct {
		slug        string
		environment string
		displayName string
	}{
		{"some-slug", "prod", "some-slug-prod"},
		{"some-slug", "production", "some-slug-production"},
		{"some-verry-unnecessarily-long-slug", "dev", "some-verry-unnecessarily-l-dev"},
		{"some-verry-unnecessarily-long-slug", "prod", "some-verry-unnecessarily-prod"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.displayName, google_gcp_reconciler.GetProjectDisplayName(slug.Slug(tt.slug), tt.environment))
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	gcpServices := &google_gcp_reconciler.GcpServices{}

	t.Run("fail early when unable to load reconciler state", func(t *testing.T) {
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)

		auditLogger.
			On("WithComponentName", types.ComponentNameGoogleGcpProject).
			Return(auditLogger).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, teamSlug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()

		err = google_gcp_reconciler.
			New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, log, nil).
			Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "load reconciler state")
	})

	t.Run("remove state when it does not refer to any projects", func(t *testing.T) {
		auditLogger := auditlogger.NewMockAuditLogger(t)
		database := db.NewMockDatabase(t)

		auditLogger.
			On("WithComponentName", types.ComponentNameGoogleGcpProject).
			Return(auditLogger).
			Once()
		database.
			On("LoadReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, teamSlug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("RemoveReconcilerStateForTeam", ctx, google_gcp_reconciler.Name, teamSlug).
			Return(nil).
			Once()

		err = google_gcp_reconciler.
			New(database, auditLogger, clusters, gcpServices, tenantName, tenantDomain, cnrmRoleName, billingAccount, log, nil).
			Delete(ctx, teamSlug, correlationID)
		assert.NoError(t, err)
	})
}
