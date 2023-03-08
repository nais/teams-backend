package google_workspace_admin_reconciler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

func TestReconcile(t *testing.T) {
	const (
		domain           = "example.com"
		gkeSecurityGroup = "gke-security-groups@example.com"
	)
	correlationID := uuid.New()

	t.Run("error when unable to load state", func(t *testing.T) {
		ctx := context.Background()

		ts := test.HttpServerWithHandlers(t, []http.HandlerFunc{})
		defer ts.Close()

		teamSlug := slug.Slug("my-team")
		input := reconcilers.Input{
			CorrelationID: correlationID,
			Team: db.Team{
				Team: &sqlc.Team{
					Slug:    teamSlug,
					Purpose: "some purpose",
				},
			},
		}

		auditLog := auditlogger.NewMockAuditLogger(t)
		log := logger.NewMockLogger(t)
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamSlug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()

		service, _ := admin_directory_v1.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(ts.URL))

		reconciler := google_workspace_admin_reconciler.New(database, auditLog, domain, service, log)
		err := reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "unable to load system state")
	})

	t.Run("empty state, create group", func(t *testing.T) {
		ctx := context.Background()

		consoleUser1 := consoleUserWithEmail("user1@example.com")
		consoleUser2 := consoleUserWithEmail("user2@example.com")
		addMe := consoleUserWithEmail("add-me@example.com")
		removeMe := consoleUserWithEmail("remove-me@example.com")

		teamSlug := slug.Slug("my-team")
		team := db.Team{
			Team: &sqlc.Team{
				Slug:    teamSlug,
				Purpose: "some purpose",
			},
		}
		existingTeamMembers := []*db.User{
			consoleUser1, // exists in the Google group
			consoleUser2, // exists in the Google group
			addMe,        // does not exist in the Google group, will be added
		}
		input := reconcilers.Input{
			CorrelationID: correlationID,
			Team:          team,
			TeamMembers:   existingTeamMembers,
		}

		expectedGoogleGroupEmail := "nais-team-my-team@example.com" // will be generated by Console
		googleGroupId := uuid.New().String()
		googleUserId1 := uuid.New().String()
		googleUserId2 := uuid.New().String()
		googleUserId4 := uuid.New().String()

		ts := test.HttpServerWithHandlers(t, []http.HandlerFunc{
			// create group
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)

				googleGroup := admin_directory_v1.Group{}
				err := json.NewDecoder(r.Body).Decode(&googleGroup)
				assert.NoError(t, err)
				assert.Equal(t, expectedGoogleGroupEmail, googleGroup.Email)

				googleGroup.Id = googleGroupId
				rsp, err := googleGroup.MarshalJSON()
				assert.NoError(t, err)

				w.Write(rsp)
			},

			// list existing members
			func(w http.ResponseWriter, r *http.Request) {
				members := admin_directory_v1.Members{
					Members: []*admin_directory_v1.Member{
						{Id: googleUserId1, Email: "user1@example.com"},     // is already a team member in Console
						{Id: googleUserId2, Email: "user2@example.com"},     // is already a team member in Console
						{Id: googleUserId4, Email: "remove-me@example.com"}, // is not a Console team member, will be removed from the Google group
					},
				}
				rsp, err := members.MarshalJSON()
				assert.NoError(t, err)
				w.Write(rsp)
			},

			// delete member
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/groups/"+googleGroupId+"/members/"+googleUserId4)
			},

			// add missing member
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)

				addedMember := admin_directory_v1.Member{}
				err := json.NewDecoder(r.Body).Decode(&addedMember)
				assert.NoError(t, err)
				assert.Equal(t, addMe.Email, addedMember.Email)

				rsp, err := addedMember.MarshalJSON()
				assert.NoError(t, err)
				w.Write(rsp)
			},

			// add to GKE security group
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)

				assert.Contains(t, r.URL.Path, "/groups/"+gkeSecurityGroup+"/members")

				addedMember := admin_directory_v1.Member{}
				err := json.NewDecoder(r.Body).Decode(&addedMember)
				assert.NoError(t, err)
				assert.Equal(t, expectedGoogleGroupEmail, addedMember.Email)

				rsp, err := addedMember.MarshalJSON()
				assert.NoError(t, err)
				w.Write(rsp)
			},
		})
		defer ts.Close()

		service, _ := admin_directory_v1.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(ts.URL))

		log := logger.NewMockLogger(t)

		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamSlug, mock.Anything).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamSlug, mock.MatchedBy(func(state reconcilers.GoogleWorkspaceState) bool {
				return *state.GroupEmail == expectedGoogleGroupEmail
			})).
			Return(nil).
			Once()
		database.
			On("GetUserByEmail", ctx, removeMe.Email).
			Return(removeMe, nil).
			Once()

		auditLog := auditlogger.NewMockAuditLogger(t)
		auditLog.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == correlationID && fields.Action == sqlc.AuditActionGoogleWorkspaceAdminCreate
			}), mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "Created Google")
			}), expectedGoogleGroupEmail).
			Return(nil).
			Once()
		auditLog.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(teamSlug) && targets[1].Identifier == removeMe.Email
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == correlationID && fields.Action == sqlc.AuditActionGoogleWorkspaceAdminDeleteMember
			}), mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "Deleted member")
			}), removeMe.Email, expectedGoogleGroupEmail).
			Return(nil).
			Once()
		auditLog.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(teamSlug) && targets[1].Identifier == addMe.Email
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == correlationID && fields.Action == sqlc.AuditActionGoogleWorkspaceAdminAddMember
			}), mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "Added member")
			}), addMe.Email, expectedGoogleGroupEmail).
			Return(nil).
			Once()
		auditLog.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(teamSlug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.CorrelationID == correlationID && fields.Action == sqlc.AuditActionGoogleWorkspaceAdminAddToGkeSecurityGroup
			}), mock.MatchedBy(func(msg string) bool {
				return strings.HasPrefix(msg, "Added group")
			}), expectedGoogleGroupEmail, gkeSecurityGroup).
			Return(nil).
			Once()

		reconciler := google_workspace_admin_reconciler.New(database, auditLog, domain, service, log)
		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})
}

func Test_Delete(t *testing.T) {
	const domain = "example.com"
	groupEmail := "my-team@example.com"
	correlationID := uuid.New()
	ctx := context.Background()
	teamSlug := slug.Slug("my-team")
	auditLogger := auditlogger.NewMockAuditLogger(t)
	log := logger.NewMockLogger(t)

	googleAdminService, close := getAdminDirectoryServiceAndClient(t, ctx, []http.HandlerFunc{})
	defer close()

	t.Run("unable to load state", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamSlug, mock.Anything).
			Return(fmt.Errorf("some error")).
			Once()

		reconciler := google_workspace_admin_reconciler.New(database, auditLogger, domain, googleAdminService, log)
		err := reconciler.Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "load reconciler state for team")
	})

	t.Run("no group email in state", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamSlug, mock.Anything).
			Return(nil).
			Once()

		reconciler := google_workspace_admin_reconciler.New(database, auditLogger, domain, googleAdminService, log)
		err := reconciler.Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "missing group email in reconciler state")
	})

	t.Run("Google API failure", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamSlug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &groupEmail
			}).
			Return(nil).
			Once()

		googleAdminService, close := getAdminDirectoryServiceAndClient(t, ctx, []http.HandlerFunc{
			// delete group failure
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/groups/my-team@example.com")
				w.WriteHeader(http.StatusBadRequest)
			},
		})
		defer close()

		reconciler := google_workspace_admin_reconciler.New(database, auditLogger, domain, googleAdminService, log)
		err := reconciler.Delete(ctx, teamSlug, correlationID)
		assert.ErrorContains(t, err, "delete Google directory group")
	})

	t.Run("successful delete", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		database.
			On("LoadReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamSlug, mock.Anything).
			Run(func(args mock.Arguments) {
				state := args.Get(3).(*reconcilers.GoogleWorkspaceState)
				state.GroupEmail = &groupEmail
			}).
			Return(nil).
			Once()
		database.
			On("RemoveReconcilerStateForTeam", ctx, google_workspace_admin_reconciler.Name, teamSlug).
			Return(nil).
			Once()

		googleAdminService, close := getAdminDirectoryServiceAndClient(t, ctx, []http.HandlerFunc{
			// delete group
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
		})
		defer close()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On(
				"Logf",
				ctx,
				database,
				mock.MatchedBy(func(targets []auditlogger.Target) bool {
					return targets[0].Type == sqlc.AuditLogsTargetTypeTeam && targets[0].Identifier == string(teamSlug)
				}),
				mock.MatchedBy(func(fields auditlogger.Fields) bool {
					return fields.CorrelationID == correlationID && fields.Action == sqlc.AuditActionGoogleWorkspaceAdminDelete
				}),
				mock.MatchedBy(func(msg string) bool {
					return strings.HasPrefix(msg, "Delete Google directory group")
				}),
				groupEmail,
			).
			Return(nil).
			Once()

		reconciler := google_workspace_admin_reconciler.New(database, auditLogger, domain, googleAdminService, log)
		assert.Nil(t, reconciler.Delete(ctx, teamSlug, correlationID))
	})
}

func consoleUserWithEmail(email string) *db.User {
	return &db.User{User: &sqlc.User{ID: uuid.New(), Email: email}}
}

func getAdminDirectoryServiceAndClient(t *testing.T, ctx context.Context, handlers []http.HandlerFunc) (service *admin_directory_v1.Service, closer func()) {
	ts := test.HttpServerWithHandlers(t, handlers)
	closer = ts.Close
	service, _ = admin_directory_v1.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(ts.URL))
	return
}
