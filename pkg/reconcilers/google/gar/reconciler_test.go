package gar_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/reconcilers/google/gar"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type fakeArtifaceRegistry struct {
	createCounter int
	create        func(ctx context.Context, r *artifactregistrypb.CreateRepositoryRequest) (*longrunningpb.Operation, error)

	getCounter int
	get        func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error)

	updateCounter int
	update        func(ctx context.Context, r *artifactregistrypb.UpdateRepositoryRequest) (*artifactregistrypb.Repository, error)

	srv *grpc.Server

	artifactregistrypb.UnimplementedArtifactRegistryServer
}

func (f *fakeArtifaceRegistry) CreateRepository(ctx context.Context, r *artifactregistrypb.CreateRepositoryRequest) (*longrunningpb.Operation, error) {
	f.createCounter++
	return f.create(ctx, r)
}

func (f *fakeArtifaceRegistry) GetRepository(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
	f.getCounter++
	return f.get(ctx, r)
}

func (f *fakeArtifaceRegistry) UpdateRepository(ctx context.Context, r *artifactregistrypb.UpdateRepositoryRequest) (*artifactregistrypb.Repository, error) {
	f.updateCounter++
	return f.update(ctx, r)
}

func (f *fakeArtifaceRegistry) assert(t *testing.T) {
	if f.create != nil {
		assert.Equal(t, f.createCounter, 1, "expected 1 call to create")
	}
	if f.update != nil {
		assert.Equal(t, f.updateCounter, 1, "expected 1 call to update")
	}
	if f.get != nil {
		assert.Equal(t, f.getCounter, 1, "expected 1 call to get")
	}
}

func (f *fakeArtifaceRegistry) start(t *testing.T, ctx context.Context) (*artifactregistry.Client, error) {
	l, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)

	f.srv = grpc.NewServer()
	artifactregistrypb.RegisterArtifactRegistryServer(f.srv, f)
	go func() {
		if err := f.srv.Serve(l); err != nil {
			panic(err)
		}
	}()

	return artifactregistry.NewClient(ctx,
		option.WithEndpoint(l.Addr().String()),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithInsecure()),
	)
}

func (f *fakeArtifaceRegistry) stop() {
	f.srv.Stop()
}

func TestReconcile(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)

	const (
		managementProjectID = "management-project-123"
		slug                = slug.Slug("team")
	)

	correlationID := uuid.New()
	team := db.Team{Team: &sqlc.Team{Slug: slug}}
	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}

	ctx := context.Background()
	repositoryName := fmt.Sprintf("projects/%s/locations/europe-north1/repositories/%s", managementProjectID, string(slug))
	repositoryDescription := fmt.Sprintf("Docker repository for team %q. Managed by NAIS Console.", string(slug))

	t.Run("fail when get repository fails", func(t *testing.T) {
		fake := &fakeArtifaceRegistry{
			get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
				return nil, fmt.Errorf("test error")
			},
		}
		client, err := fake.start(t, ctx)
		assert.NoError(t, err)
		defer fake.stop()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := gar.New(auditLogger, managementProjectID, client, log)
		err = reconciler.Reconcile(ctx, input)
		assert.ErrorContains(t, err, "test error")
	})

	t.Run("create when no repository already exists", func(t *testing.T) {
		fake := &fakeArtifaceRegistry{
			get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
				return nil, status.Errorf(codes.NotFound, "test error")
			},
			create: func(ctx context.Context, r *artifactregistrypb.CreateRepositoryRequest) (*longrunningpb.Operation, error) {
				assert.Equal(t, repositoryName, r.Repository.Name)
				assert.Equal(t, repositoryDescription, r.Repository.Description)
				assert.Equal(t, artifactregistrypb.Repository_DOCKER, r.Repository.Format)
				assert.Equal(t, string(slug), r.Repository.Labels["team"])

				payload := anypb.Any{}
				err := anypb.MarshalFrom(&payload, r.Repository, proto.MarshalOptions{})
				assert.NoError(t, err)

				return &longrunningpb.Operation{
					Done: true,
					Result: &longrunningpb.Operation_Response{
						Response: &payload,
					},
				}, nil
			},
		}

		client, err := fake.start(t, ctx)
		assert.NoError(t, err)
		defer fake.stop()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := gar.New(auditLogger, managementProjectID, client, log)
		err = reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
		fake.assert(t)
	})

	t.Run("no update when existing repository is up to date", func(t *testing.T) {
		repo := &artifactregistrypb.Repository{
			Name:        repositoryName,
			Format:      artifactregistrypb.Repository_DOCKER,
			Description: repositoryDescription,
			Labels:      map[string]string{"team": string(slug)},
		}
		fake := &fakeArtifaceRegistry{
			get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
				return repo, nil
			},
		}

		client, err := fake.start(t, ctx)
		assert.NoError(t, err)
		defer fake.stop()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := gar.New(auditLogger, managementProjectID, client, log)
		err = reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
		fake.assert(t)
	})

	t.Run("update when existing repository has incorrect desc", func(t *testing.T) {
		repo := &artifactregistrypb.Repository{
			Name:        repositoryName,
			Format:      artifactregistrypb.Repository_DOCKER,
			Description: repositoryDescription + "invalid",
			Labels:      map[string]string{"team": string(slug)},
		}

		fake := &fakeArtifaceRegistry{
			get: func(ctx context.Context, r *artifactregistrypb.GetRepositoryRequest) (*artifactregistrypb.Repository, error) {
				return repo, nil
			},
			update: func(ctx context.Context, r *artifactregistrypb.UpdateRepositoryRequest) (*artifactregistrypb.Repository, error) {
				assert.Len(t, r.UpdateMask.Paths, 1)
				r.Repository.Description = repositoryDescription

				return r.Repository, nil
			},
		}

		client, err := fake.start(t, ctx)
		assert.NoError(t, err)
		defer fake.stop()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler := gar.New(auditLogger, managementProjectID, client, log)
		err = reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
		fake.assert(t)
	})
}
