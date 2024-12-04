// A generated module for LinkFetcher functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/link-fetcher/internal/dagger"
)

type LinkFetcher struct{}

// Internal method to return a Golang container with the application inside
func (m *LinkFetcher) golang(ctx context.Context, src *dagger.Directory) *dagger.Container {
	return dag.Container().
		From("golang:latest").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0")
}

// Format the source code
func (m *LinkFetcher) Fmt(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!*.go", "!go.mod", "!go.sum"]
	src *dagger.Directory,
) *dagger.Directory {
	return m.golang(ctx, src).WithExec([]string{"go", "fmt", "."}).Directory("/src")
}

// Test the application by running it on https://news.ycombinator.com/
func (m *LinkFetcher) Test(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!*.go", "!go.mod", "!go.sum"]
	src *dagger.Directory,
) (string, error) {
	return m.golang(ctx, src).WithExec([]string{"go", "run", ".", "https://news.ycombinator.com/"}).Stdout(ctx)
}

// Run the linter
func (m *LinkFetcher) Lint(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!*.go", "!go.mod", "!go.sum"]
	src *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("golangci/golangci-lint").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{"go", "mod", "tidy"}).
		WithExec([]string{"golangci-lint", "run", "."}).
		Stdout(ctx)
}

// Scan the given image for vulnerabilities
func (m *LinkFetcher) Scan(ctx context.Context, image *string) (string, error) {
	return dag.Container().
		From("aquasec/trivy").
		WithExec([]string{"trivy", "image", *image}).
		Stdout(ctx)
}

// Build a Docker image and publish it on ttl.sh
// Takend from https://docs.dagger.io/cookbook/#perform-a-multi-stage-build
func (m *LinkFetcher) Build(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!*.go", "!go.mod", "!go.sum"]
	src *dagger.Directory,
) (string, error) {
	builder := m.golang(ctx, src).WithExec([]string{"go", "build", "-o", "link-fetcher"})

	img := dag.Container().
		From("alpine").
		WithExec([]string{"apk", "update"}).
		WithExec([]string{"apk", "upgrade"}).
		WithExec([]string{"rm", "-rf", "/var/cache/apk/*"}).
		WithFile("/bin/link-fetcher", builder.File("/src/link-fetcher")).
		WithExec([]string{"addgroup", "-g", "1000", "-S", "app"}).
		WithExec([]string{"adduser", "-u", "1000", "-S", "app", "-G", "app"}).
		WithUser("app").
		WithEntrypoint([]string{"/bin/link-fetcher"})

	addr, err := img.Publish(ctx, "ttl.sh/link-fetcher:latest")
	if err != nil {
		return "", err
	}

	return addr, nil
}
