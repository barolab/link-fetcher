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
	"encoding/json"
	"fmt"
	"time"

	"dagger/link-fetcher/internal/dagger"
)

type LinkFetcher struct {
	kc *dagger.Container
}

// Pretty print a map of strings
//
// Examples:
//
//	  pretty(map[string]string{"key1": "value1", "key2", "value2"})
//
//		 - key1:
//	  value1
//
//	  - key2
//	  value2
func pretty(outputs map[string]string) string {
	result := ""
	for k, v := range outputs {
		result += fmt.Sprintf("- %s:\n%s\n\n", k, v)
	}
	return result
}

// Internal method to return a Golang container with the application inside
func (m *LinkFetcher) golang(src *dagger.Directory) *dagger.Container {
	return dag.Container().
		From("golang:latest").
		WithDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0")
}

// Internal method to return a kubectl container so we can deploy our application and test it
func (m *LinkFetcher) kubectl(src *dagger.Directory, kubeconfig *dagger.File) *dagger.Container {
	if m.kc == nil {
		m.kc = dag.Container().
			From("bitnami/kubectl").
			WithoutEntrypoint().
			WithEnvVariable("KUBECONFIG", "/.kube/config").
			WithFile("/.kube/config", kubeconfig, dagger.ContainerWithFileOpts{Permissions: 1001}).
			WithUser("1001").
			WithDirectory("/src", src)
	}

	return m.kc
}

// Format the source code
func (m *LinkFetcher) Fmt(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!*.go", "!go.mod", "!go.sum"]
	src *dagger.Directory,
) *dagger.Directory {
	return m.golang(src).WithExec([]string{"go", "fmt", "."}).Directory("/src")
}

// Run the application with the following argument: https://news.ycombinator.com/
func (m *LinkFetcher) Run(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!*.go", "!go.mod", "!go.sum"]
	src *dagger.Directory,
) (string, error) {
	return m.golang(src).WithExec([]string{"go", "run", ".", "https://news.ycombinator.com/"}).Stdout(ctx)
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
	builder := m.golang(src).WithExec([]string{"go", "build", "-o", "link-fetcher"})

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

// Deploy our application to a Kubernetes cluster
func (m *LinkFetcher) Deploy(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!**/*.yaml"]
	src *dagger.Directory,
	// +defaultPath="$HOME/.kube/config"
	kubeconfig *dagger.File,
) (string, error) {
	return m.kubectl(src, kubeconfig).WithExec([]string{"sh", "-c", "kubectl apply -f /src/kubernetes/deployment.yaml"}).Stdout(ctx)
}

// Validate the application is working as expected
func (m *LinkFetcher) Validate(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!*.go", "!go.mod", "!go.sum", "!**/*.yaml"]
	src *dagger.Directory,
	// +defaultPath="$HOME/.kube/config"
	kubeconfig *dagger.File,
) (string, error) {
	kubectl := m.kubectl(src, kubeconfig)

	// Wait for the POD to boot
	_, err := kubectl.WithExec([]string{"sh", "-c", "kubectl rollout status deploy link-fetcher --timeout=30s"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for link-fetcher to be ready, %w", err)
	}

	// Get the Deployment logs (it should be a JSON string)
	logs, err := kubectl.WithExec([]string{"sh", "-c", "kubectl logs deploy/link-fetcher"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs from link-fetcher, %w", err)
	}

	// Unmarshal the output of link-fetcher
	var result map[string][]string
	err = json.Unmarshal([]byte(logs), &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal the link-fetcher output, %w (got %s)", err, logs)
	}

	// Fail if link-fetcher didn't found anything
	resultCount := len(result)
	if resultCount == 0 {
		return "", fmt.Errorf("result is empty, expecting at least 1 result from link-fetcher (got %s)", logs)
	}

	return fmt.Sprintf("Found %d results, validation succeeded", resultCount), nil
}

// Run a full integration test by deploying our application to K3S
func (m *LinkFetcher) IntegrationTest(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!*.go", "!go.mod", "!go.sum", "!**/*.yaml"]
	src *dagger.Directory,
) (string, error) {
	output := map[string]string{}

	// Build the image to make sure it's available
	build, err := m.Build(ctx, src)
	output["build"] = build
	if err != nil {
		return "", err
	}

	// Start a K8S cluster in a container using K3S
	k3s := dag.K3S("test")
	kServer := k3s.Server()
	kServer, err = kServer.Start(ctx)
	if err != nil {
		return "", err
	}
	defer kServer.Stop(ctx)

	// Deploy our application
	deploy, err := m.Deploy(ctx, src, k3s.Config())
	output["deploy"] = deploy
	if err != nil {
		return "", err
	}

	// Wait 10 seconds for the POD to boot and add a timeout to context to ensure we don't enter an infine loop
	time.Sleep(10 * time.Second)

	// Validate the application is running
	validate, err := m.Validate(ctx, src, k3s.Config())
	output["validate"] = validate
	if err != nil {
		return "", err
	}

	return pretty(output), nil
}
