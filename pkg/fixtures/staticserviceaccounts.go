package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nais/console/pkg/serviceaccount"
	"strings"

	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/slug"
)

type ServiceAccount struct {
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	APIKey string   `json:"apiKey"`
}

const NaisServiceAccountPrefix = "nais-"

func parseAndValidateServiceAccounts(serviceAccountsJson string) ([]ServiceAccount, error) {
	serviceAccounts := make([]ServiceAccount, 0)
	err := json.NewDecoder(strings.NewReader(serviceAccountsJson)).Decode(&serviceAccounts)
	if err != nil {
		return nil, err
	}

	for _, serviceAccount := range serviceAccounts {
		if !strings.HasPrefix(serviceAccount.Name, NaisServiceAccountPrefix) {
			return nil, fmt.Errorf("service account is missing required '%s' prefix: '%s'", NaisServiceAccountPrefix, serviceAccount.Name)
		}

		if len(serviceAccount.Roles) == 0 {
			return nil, fmt.Errorf("service account must have at least one role: '%s'", serviceAccount.Name)
		}

		if serviceAccount.APIKey == "" {
			return nil, fmt.Errorf("service account is missing an API key: '%s'", serviceAccount.Name)
		}

		for _, role := range serviceAccount.Roles {
			if !sqlc.RoleName(role).Valid() {
				return nil, fmt.Errorf("invalid role name: '%s' for service account '%s'", role, serviceAccount.Name)
			}
		}
	}

	return serviceAccounts, nil
}

// SetupStaticServiceAccounts Create a set of service accounts with roles and API keys
func SetupStaticServiceAccounts(ctx context.Context, database db.Database, serviceAccountsJson, tenantDomain string) error {
	serviceAccounts, err := parseAndValidateServiceAccounts(serviceAccountsJson)
	if err != nil {
		return err
	}

	return database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		for _, serviceAccountFromInput := range serviceAccounts {
			email := serviceaccount.Email(slug.Slug(serviceAccountFromInput.Name), tenantDomain)

			serviceAccount, err := dbtx.GetUserByEmail(ctx, email)
			if err != nil {
				serviceAccount, err = dbtx.AddUser(ctx, serviceAccountFromInput.Name, email)
				if err != nil {
					return err
				}
			}

			err = dbtx.RemoveUserRoles(ctx, serviceAccount.ID)
			if err != nil {
				return err
			}

			err = dbtx.RemoveApiKeysFromUser(ctx, serviceAccount.ID)
			if err != nil {
				return err
			}

			for _, roleName := range serviceAccountFromInput.Roles {
				err = dbtx.AssignGlobalRoleToUser(ctx, serviceAccount.ID, sqlc.RoleName(roleName))
				if err != nil {
					return err
				}
			}

			err = dbtx.CreateAPIKey(ctx, serviceAccountFromInput.APIKey, serviceAccount.ID)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
