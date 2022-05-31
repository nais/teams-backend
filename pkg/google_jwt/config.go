package google_jwt

import (
	"fmt"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	admin_directory_v1 "google.golang.org/api/admin/directory/v1"
	"io/ioutil"
)

func GetConfig(credentialsFile, delegatedUser string) (*jwt.Config, error) {
	b, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read google credentials file: %w", err)
	}

	cf, err := google.JWTConfigFromJSON(
		b,
		admin_directory_v1.AdminDirectoryUserReadonlyScope,
		admin_directory_v1.AdminDirectoryGroupScope,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize google credentials: %w", err)
	}

	cf.Subject = delegatedUser

	return cf, nil
}
