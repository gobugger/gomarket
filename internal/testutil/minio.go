package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// MinioContainer holds the running container and a ready‑to‑use client.
type MinioContainer struct {
	Container testcontainers.Container
	Endpoint  string // e.g. "127.0.0.1:32768"
	Client    *minio.Client
	AccessKey string
	SecretKey string
}

// NewMinioContainer starts a MinIO Docker container and returns a helper
// that can be used in any test. The caller is responsible for calling
// c.Terminate(ctx) (usually via defer) when the test finishes.
func NewMinioContainer(ctx context.Context) (*MinioContainer, error) {
	const (
		image         = "minio/minio:latest"
		accessKey     = "admin"
		secretKey     = "supersecret"
		containerPort = "9000/tcp"
	)

	// -----------------------------------------------------------------
	// 1️⃣ Define the container request
	// -----------------------------------------------------------------
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{containerPort},
		Env: map[string]string{
			"MINIO_ROOT_USER":     accessKey,
			"MINIO_ROOT_PASSWORD": secretKey,
		},
		Cmd:        []string{"server", "/data"},
		WaitingFor: wait.ForListeningPort(containerPort).WithStartupTimeout(10 * time.Minute),
	}
	// -----------------------------------------------------------------
	// 2️⃣ Start the container
	// -----------------------------------------------------------------
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start minio container: %w", err)
	}

	// -----------------------------------------------------------------
	// 3️⃣ Resolve host & mapped port
	// -----------------------------------------------------------------
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}
	mappedPort, err := container.MappedPort(ctx, containerPort)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}
	endpoint := fmt.Sprintf("%s:%s", host, mappedPort.Port())

	// -----------------------------------------------------------------
	// 4️⃣ Build a MinIO client (plain HTTP – container has no TLS)
	// -----------------------------------------------------------------
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	buckets := []string{
		"product",
		"thumbnail",
		"logo",
		"inventory",
	}

	for _, bucket := range buckets {
		if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	// -----------------------------------------------------------------
	// 5️⃣ Return the helper struct
	// -----------------------------------------------------------------
	return &MinioContainer{
		Container: container,
		Endpoint:  endpoint,
		Client:    minioClient,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}, nil
}

// Terminate stops the container. Call it in a defer right after NewMinioContainer.
func (c *MinioContainer) Terminate(ctx context.Context) error {
	return c.Container.Terminate(ctx)
}

// ---------------------------------------------------------------------
// Convenience: create a bucket that is guaranteed to be unique per test.
// ---------------------------------------------------------------------
func (c *MinioContainer) CreateTempBucket(ctx context.Context) (string, error) {
	bucket := fmt.Sprintf("test-%d", time.Now().UnixNano())
	if err := c.Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
		return "", fmt.Errorf("make bucket %s: %w", bucket, err)
	}
	return bucket, nil
}
