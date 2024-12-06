package main

import (
	"context"
	"dagger/link-fetcher/internal/dagger"
	"encoding/json"
	"fmt"
	"time"
)

// Internal method to return a kubectl container
func (m *LinkFetcher) kubectl(
	src *dagger.Directory,
	kubeconfig *dagger.File,
	kube *dagger.Service,
	kubeAddr string,
	kubePort string,
	certs *dagger.Directory,
	certsPath string,
) *dagger.Container {
	if m.kc == nil {
		m.kc = dag.Container().
			From("bitnami/kubectl").
			WithoutEntrypoint().
			WithEnvVariable("KUBECONFIG", "/.kube/config").
			WithFile("/.kube/config", kubeconfig, dagger.ContainerWithFileOpts{Owner: "1001", Permissions: 0600}).
			WithUser("1001").
			WithDirectory("/src", src)
	}

	// If a host service is given, mount it in the kubectl container
	// and replace the address in the kubeconfig file with the service name
	if kube != nil {
		replace := fmt.Sprintf(`s/https:.*:%s/https:\/\/%s:%s/g`, kubePort, kubeAddr, kubePort)
		m.kc = m.kc.
			WithServiceBinding(kubeAddr, kube).
			WithExec([]string{"sed", "-i", replace, "/.kube/config"})
	}

	if certs != nil {
		m.kc = m.kc.WithDirectory(certsPath, certs, dagger.ContainerWithDirectoryOpts{Owner: "1001"})
	}

	return m.kc
}

func (m *LinkFetcher) Kubectl(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!**/*.yaml"]
	src *dagger.Directory,
	kubeconfig *dagger.File,
	// +optional
	kube *dagger.Service,
	// +optional
	// +default="kube"
	kubeAddr string,
	// +optional
	// +default="443"
	kubePort string,
	// +optional
	// +ignore=["*", "!**/*.crt", "!**/*.key"]
	certs *dagger.Directory,
	// +optional
	certsPath string,
	// +default=false
	insecure bool,
	args string,
) (string, error) {
	return m.kubectl(src, kubeconfig, kube, kubeAddr, kubePort, certs, certsPath).
		WithExec([]string{"sh", "-c", fmt.Sprintf("kubectl --insecure-skip-tls-verify=%t %s", insecure, args)}).
		Stdout(ctx)
}

// Deploy our application to a Kubernetes cluster
func (m *LinkFetcher) Deploy(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!**/*.yaml"]
	src *dagger.Directory,
	kubeconfig *dagger.File,
	// +optional
	kube *dagger.Service,
	// +optional
	// +default="kube"
	kubeAddr string,
	// +optional
	// +default="443"
	kubePort string,
	// +optional
	// +ignore=["*", "!**/*.crt", "!**/*.key"]
	certs *dagger.Directory,
	// +optional
	certsPath string,
	// +default=false
	insecure bool,
) (string, error) {
	return m.Kubectl(ctx, src, kubeconfig, kube, kubeAddr, kubePort, certs, certsPath, insecure, "apply -f /src/kubernetes/deployment.yaml")
}

// Deploy our application to a Kubernetes cluster
func (m *LinkFetcher) Status(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!**/*.yaml"]
	src *dagger.Directory,
	kubeconfig *dagger.File,
	// +optional
	kube *dagger.Service,
	// +optional
	// +default="kube"
	kubeAddr string,
	// +optional
	// +default="443"
	kubePort string,
	// +optional
	// +ignore=["*", "!**/*.crt", "!**/*.key"]
	certs *dagger.Directory,
	// +optional
	certsPath string,
	// +default=false
	insecure bool,
	// +default="60s"
	timeout string,
) (string, error) {
	return m.Kubectl(ctx, src, kubeconfig, kube, kubeAddr, kubePort, certs, certsPath, insecure, fmt.Sprintf("rollout status deploy link-fetcher --timeout=%s", timeout))
}

// Deploy our application to a Kubernetes cluster
func (m *LinkFetcher) Logs(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!**/*.yaml"]
	src *dagger.Directory,
	kubeconfig *dagger.File,
	// +optional
	kube *dagger.Service,
	// +optional
	// +default="kube"
	kubeAddr string,
	// +optional
	// +default="443"
	kubePort string,
	// +optional
	// +ignore=["*", "!**/*.crt", "!**/*.key"]
	certs *dagger.Directory,
	// +optional
	certsPath string,
	// +default=false
	insecure bool,
) (string, error) {
	return m.Kubectl(ctx, src, kubeconfig, kube, kubeAddr, kubePort, certs, certsPath, insecure, "logs deploy/link-fetcher")
}

// Deploy our application to a Kubernetes cluster
func (m *LinkFetcher) Validate(
	ctx context.Context,
	// +defaultPath="./"
	// +ignore=["*", "!**/*.yaml"]
	src *dagger.Directory,
	kubeconfig *dagger.File,
	// +optional
	kube *dagger.Service,
	// +optional
	// +default="kube"
	kubeAddr string,
	// +optional
	// +default="443"
	kubePort string,
	// +optional
	// +ignore=["*", "!**/*.crt", "!**/*.key"]
	certs *dagger.Directory,
	// +optional
	certsPath string,
	// +default=false
	insecure bool,
	// +default="60s"
	timeout string,
) (string, error) {
	_, err := m.Status(ctx, src, kubeconfig, kube, kubeAddr, kubePort, certs, certsPath, insecure, timeout)
	if err != nil {
		return "", err
	}

	logs, err := m.Logs(ctx, src, kubeconfig, kube, kubeAddr, kubePort, certs, certsPath, insecure)
	if err != nil {
		return "", err
	}

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

	// Wait a bit for the K3S cluster to become ready
	time.Sleep(1 * time.Second)

	// Deploy our application
	deploy, err := m.Deploy(ctx, src, k3s.Config(), nil, "", "", nil, "", false)
	output["deploy"] = deploy
	if err != nil {
		return "", err
	}

	// Validate the application is running
	validate, err := m.Validate(ctx, src, k3s.Config(), nil, "", "", nil, "", false, "60s")
	output["validate"] = validate
	if err != nil {
		return "", err
	}

	return pretty(output), nil
}
