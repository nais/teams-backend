package gcp_team_reconciler_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

func TestTheWholeThing(t *testing.T) {

	ctx := context.Background()
	b, err := ioutil.ReadFile("/tmp/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	creds, err := google.CredentialsFromJSON(ctx, b, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := oauth2.NewClient(ctx, creds.TokenSource)

	srv, err := admin.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve directory Client %v", err)
	}

	r, err := srv.Users.List().Customer("customer").MaxResults(10).
		OrderBy("email").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve users in domain: %v", err)
	}

	if len(r.Users) == 0 {
		fmt.Print("No users found.\n")
	} else {
		fmt.Print("Users:\n")
		for _, u := range r.Users {
			fmt.Printf("%s (%s)\n", u.PrimaryEmail, u.Name.FullName)
		}
	}
}
