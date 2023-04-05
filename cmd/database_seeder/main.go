package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	databaseConnectRetries = 5
)

type seedConfig struct {
	DatabaseURL       string `envconfig:"CONSOLE_DATABASE_URL" default:"postgres://console:console@localhost:3002/console?sslmode=disable"`
	Domain            string `envconfig:"CONSOLE_TENANT_DOMAIN" default:"example.com"`
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

	database, err := setupDatabase(ctx, cfg.DatabaseURL, log)
	if err != nil {
		return err
	}

	if !*cfg.ForceSeed {
		if existingUsers, err := database.GetUsers(ctx); len(existingUsers) != 0 || err != nil {
			return fmt.Errorf("database already has users, abort")
		}

		if existingTeams, err := database.GetTeams(ctx); len(existingTeams) != 0 || err != nil {
			return fmt.Errorf("database already has teams, abort")
		}
	}

	err = database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		emails := map[string]struct{}{}
		users := make([]*db.User, 0)
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

		slugs := map[string]struct{}{}
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
				_ = dbtx.SetTeamMemberRole(ctx, users[rand.Intn(usersCreated)].ID, team.Slug, sqlc.RoleNameTeamowner)
			}

			for o := 0; o < *cfg.NumMembersPerTeam; o++ {
				_ = dbtx.SetTeamMemberRole(ctx, users[rand.Intn(usersCreated)].ID, team.Slug, sqlc.RoleNameTeammember)
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

func setupDatabase(ctx context.Context, dbUrl string, log logger.Logger) (db.Database, error) {
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, err
	}

	var dbc *pgxpool.Pool
	for i := 0; i < databaseConnectRetries; i++ {
		dbc, err = pgxpool.ConnectConfig(ctx, config)
		if err == nil {
			break
		}

		log.Warnf("unable to connect to the database: %s", err)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	if dbc == nil {
		return nil, fmt.Errorf("giving up connecting to the database after %d attempts: %w", databaseConnectRetries, err)
	}

	err = db.Migrate(dbc.Config().ConnString())
	if err != nil {
		return nil, err
	}

	queries := db.Wrap(sqlc.New(dbc), dbc)
	return db.NewDatabase(queries), nil
}
