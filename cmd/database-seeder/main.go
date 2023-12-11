package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/pkg/usersync"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type seedConfig struct {
	DatabaseURL       string `envconfig:"TEAMS_BACKEND_DATABASE_URL" default:"postgres://console:console@localhost:3002/console?sslmode=disable"`
	Domain            string `envconfig:"TEAMS_BACKEND_TENANT_DOMAIN" default:"example.com"`
	NumUsers          *int
	NumTeams          *int
	NumOwnersPerTeam  *int
	NumMembersPerTeam *int
	ForceSeed         *bool
}

func newSeedConfig() (*seedConfig, error) {
	cfg := &seedConfig{}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	cfg.NumUsers = flag.Int("users", 1000, "number of users to insert")
	cfg.NumTeams = flag.Int("teams", 200, "number of teams to insert")
	cfg.NumOwnersPerTeam = flag.Int("owners", 3, "number of owners per team")
	cfg.NumMembersPerTeam = flag.Int("members", 10, "number of members per team")
	cfg.ForceSeed = flag.Bool("force", false, "seed regardless of existing database content")
	flag.Parse()

	return cfg, nil
}

func main() {
	cfg, err := newSeedConfig()
	if err != nil {
		fmt.Printf("fatal: %s", err)
		os.Exit(1)
	}

	log, err := logger.GetLogger("text", "INFO")
	if err != nil {
		fmt.Printf("fatal: %s", err)
		os.Exit(2)
	}

	err = run(cfg, log)
	if err != nil {
		log.WithError(err).Error("fatal error in run()")
		os.Exit(3)
	}
}

func run(cfg *seedConfig, log logger.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	firstNames, err := fileToSlice("data/first_names.txt")
	if err != nil {
		return err
	}
	numFirstNames := len(firstNames)

	lastNames, err := fileToSlice("data/last_names.txt")
	if err != nil {
		return err
	}
	numLastNames := len(lastNames)

	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}

	emails := map[string]struct{}{}
	slugs := map[string]struct{}{}

	if !*cfg.ForceSeed {
		if existingUsers, err := database.GetUsers(ctx, nil, nil); len(existingUsers) != 0 || err != nil {
			return fmt.Errorf("database already has users, abort")
		}

		if existingTeams, err := database.GetTeams(ctx, nil, nil); len(existingTeams) != 0 || err != nil {
			return fmt.Errorf("database already has teams, abort")
		}
	} else {
		users, err := database.GetUsers(ctx, nil, nil)
		if err != nil {
			return err
		}
		for _, user := range users {
			emails[user.Email] = struct{}{}
		}

		teams, err := database.GetTeams(ctx, nil, nil)
		if err != nil {
			return err
		}
		for _, team := range teams {
			slugs[string(team.Slug)] = struct{}{}
		}
	}

	err = database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		devUser, err := dbtx.CreateUser(ctx, "dev usersen", "dev.usersen@nais.io", uuid.New().String())
		if err != nil {
			return err
		}
		adminUser, err := dbtx.CreateUser(ctx, "admin usersen", "admin.usersen@nais.io", uuid.New().String())
		if err != nil {
			return err
		}
		if err = dbtx.AssignGlobalRoleToUser(ctx, adminUser.ID, sqlc.RoleNameAdmin); err != nil {
			return err
		}
		for _, roleName := range usersync.DefaultRoleNames {
			err = dbtx.AssignGlobalRoleToUser(ctx, devUser.ID, roleName)
			if err != nil {
				return fmt.Errorf("attach default role %q to user %q: %w", roleName, devUser.Email, err)
			}
		}
		users := []*db.User{devUser}
		for i := 1; i <= *cfg.NumUsers; i++ {
			firstName := firstNames[rand.Intn(numFirstNames)]
			lastName := lastNames[rand.Intn(numLastNames)]
			name := firstName + " " + lastName
			email := nameToEmail(name, cfg.Domain)
			if _, exists := emails[email]; exists {
				continue
			}

			user, err := dbtx.CreateUser(ctx, name, email, uuid.New().String())
			if err != nil {
				return err
			}

			log.Infof("%d/%d users created", i, *cfg.NumUsers)
			users = append(users, user)
			emails[email] = struct{}{}
		}
		usersCreated := len(users)

		devteam, err := dbtx.CreateTeam(ctx, slug.Slug("devteam"), "dev-purpose", "#devteam")
		if err != nil {
			return err
		}
		err = dbtx.SetTeamMemberRole(ctx, devUser.ID, devteam.Slug, sqlc.RoleNameTeamowner)
		if err != nil {
			return err
		}

		for i := 1; i <= *cfg.NumTeams; i++ {
			name := teamName()
			if _, exists := slugs[name]; exists {
				continue
			}

			team, err := dbtx.CreateTeam(ctx, slug.Slug(name), "some purpose", "#"+name)
			if err != nil {
				return err
			}

			for o := 0; o < *cfg.NumOwnersPerTeam; o++ {
				err = dbtx.SetTeamMemberRole(ctx, users[rand.Intn(usersCreated)].ID, team.Slug, sqlc.RoleNameTeamowner)
				if err != nil {
					return err
				}
			}

			for o := 0; o < *cfg.NumMembersPerTeam; o++ {
				err = dbtx.SetTeamMemberRole(ctx, users[rand.Intn(usersCreated)].ID, team.Slug, sqlc.RoleNameTeammember)
				if err != nil {
					return err
				}
			}

			log.Infof("%d/%d teams created", i, *cfg.NumTeams)
			slugs[name] = struct{}{}
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Infof("done")
	return nil
}

func teamName() string {
	letters := []byte("abcdefghijklmnopqrstuvwxyz")
	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func nameToEmail(name, domain string) string {
	name = strings.NewReplacer(" ", ".", "æ", "ae", "ø", "oe", "å", "aa").Replace(strings.ToLower(name))
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	name, _, _ = transform.String(t, name)
	return name + "@" + domain
}

func fileToSlice(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}
