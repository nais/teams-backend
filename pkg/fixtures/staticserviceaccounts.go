package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"
)

type ServiceAccount struct {
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	APIKey string   `json:"apiKey"`
}

const NaisServiceAccountPrefix = "nais-"

type ServiceAccounts []ServiceAccount

func (s *ServiceAccounts) Decode(value string) error {
	if value == "" {
		return nil
	}

	serviceAccounts := make(ServiceAccounts, 0)
	err := json.NewDecoder(strings.NewReader(value)).Decode(&serviceAccounts)
	if err != nil {
		return err
	}

	for _, serviceAccount := range serviceAccounts {
		if !strings.HasPrefix(serviceAccount.Name, NaisServiceAccountPrefix) {
			return fmt.Errorf("service account is missing required %q prefix: %q", NaisServiceAccountPrefix, serviceAccount.Name)
		}

		if len(serviceAccount.Roles) == 0 {
			return fmt.Errorf("service account must have at least one role: %q", serviceAccount.Name)
		}

		if serviceAccount.APIKey == "" {
			return fmt.Errorf("service account is missing an API key: %q", serviceAccount.Name)
		}

		for _, role := range serviceAccount.Roles {
			if !sqlc.RoleName(role).Valid() {
				return fmt.Errorf("invalid role name: %q for service account %q", role, serviceAccount.Name)
			}
		}
	}

	*s = serviceAccounts
	return nil
}

// SetupStaticServiceAccounts Create a set of service accounts with roles and API keys
func SetupStaticServiceAccounts(ctx context.Context, database db.Database, serviceAccounts ServiceAccounts) error {
	return database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		serviceAccountNames := make(map[string]struct{})
		for _, serviceAccountFromInput := range serviceAccounts {
			serviceAccountNames[serviceAccountFromInput.Name] = struct{}{}
			serviceAccount, err := dbtx.GetServiceAccountByName(ctx, serviceAccountFromInput.Name)
			if err != nil {
				serviceAccount, err = dbtx.CreateServiceAccount(ctx, serviceAccountFromInput.Name)
				if err != nil {
					return err
				}
			}

			err = dbtx.RemoveAllServiceAccountRoles(ctx, serviceAccount.ID)
			if err != nil {
				return err
			}

			err = dbtx.RemoveApiKeysFromServiceAccount(ctx, serviceAccount.ID)
			if err != nil {
				return err
			}

			for _, roleName := range serviceAccountFromInput.Roles {
				err = dbtx.AssignGlobalRoleToServiceAccount(ctx, serviceAccount.ID, sqlc.RoleName(roleName))
				if err != nil {
					return err
				}
			}

			err = dbtx.CreateAPIKey(ctx, serviceAccountFromInput.APIKey, serviceAccount.ID)
			if err != nil {
				return err
			}
		}

		// remove all NAIS service accounts that is not present in the JSON input
		serviceAccounts, err := dbtx.GetServiceAccounts(ctx)
		if err != nil {
			return err
		}

		for _, serviceAccount := range serviceAccounts {
			if !strings.HasPrefix(serviceAccount.Name, NaisServiceAccountPrefix) {
				continue
			}

			if _, shouldExist := serviceAccountNames[serviceAccount.Name]; shouldExist {
				continue
			}

			if err := dbtx.DeleteServiceAccount(ctx, serviceAccount.ID); err != nil {
				return err
			}
		}

		return nil
	})
}
