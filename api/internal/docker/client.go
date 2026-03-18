package docker

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
)

type SandboxCreateRequest struct {
	ImageName     string
	ContainerName string
	Hostname      string
}

type SandboxContainer struct {
	ID   string
	Name string
	URL  string
}

type Client interface {
	EnsureImage(ctx context.Context, imageName string) error
	RemoveImage(ctx context.Context, imageName string) error
	CreateContainer(ctx context.Context, request SandboxCreateRequest) (*SandboxContainer, error)
	DeleteContainer(ctx context.Context, containerID string) error
	CommitContainer(ctx context.Context, containerID, targetImage string) error
}

type DockerClient struct {
	client     *client.Client
	sandboxCfg config.SandboxConfig
	dockerCfg  config.DockerConfig
}

func NewClient(sandboxCfg config.SandboxConfig, dockerCfg config.DockerConfig) (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	if _, err := cli.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("docker daemon not reachable: %w", err)
	}

	return &DockerClient{
		client:     cli,
		sandboxCfg: sandboxCfg,
		dockerCfg:  dockerCfg,
	}, nil
}

func (c *DockerClient) EnsureImage(ctx context.Context, imageName string) error {
	if imageName == "" {
		return fmt.Errorf("invalid image reference")
	}

	// Reuse an already available image locally to avoid unnecessary pulls on
	// every sandbox start.
	if _, _, err := c.client.ImageInspectWithRaw(ctx, imageName); err == nil {
		return nil
	}

	// Pulling here keeps image creation and sandbox creation idempotent from the
	// caller's point of view.
	reader, err := c.client.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("consume image pull output for %s: %w", imageName, err)
	}

	return nil
}

func (c *DockerClient) RemoveImage(ctx context.Context, imageName string) error {
	if imageName == "" {
		return fmt.Errorf("invalid image reference")
	}

	if _, err := c.client.ImageRemove(ctx, imageName, types.ImageRemoveOptions{Force: false, PruneChildren: false}); err != nil {
		return fmt.Errorf("remove image %s: %w", imageName, err)
	}

	return nil
}

func (c *DockerClient) CreateContainer(ctx context.Context, request SandboxCreateRequest) (*SandboxContainer, error) {
	if err := c.EnsureImage(ctx, request.ImageName); err != nil {
		return nil, err
	}

	// Labels are the contract with Traefik, so the API owns them in one place
	// instead of spreading routing rules across higher layers.
	containerConfig := &container.Config{
		Image:  request.ImageName,
		Labels: c.buildTraefikLabels(request.ContainerName, request.Hostname),
		Env: []string{
			"TRUSTED_PROXIES=" + c.dockerCfg.TrustedProxies,
			"SHOP_DOMAIN=" + request.Hostname,
		},
	}

	var networkingConfig *network.NetworkingConfig
	if c.dockerCfg.Network != "" {
		// The sandbox container must join the same network Traefik is watching so
		// host-based routing can resolve it immediately.
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				c.dockerCfg.Network: {},
			},
		}
	}

	resp, err := c.client.ContainerCreate(ctx, containerConfig, nil, networkingConfig, nil, request.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("create container %s: %w", request.ContainerName, err)
	}

	if err := c.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("start container %s: %w", resp.ID, err)
	}

	return &SandboxContainer{
		ID:   resp.ID,
		Name: request.ContainerName,
		URL:  "https://" + request.Hostname,
	}, nil
}

func (c *DockerClient) DeleteContainer(ctx context.Context, containerID string) error {
	timeout := 0
	// Force removal is intentional because expired demo sandboxes should not
	// block cleanup on graceful shutdown behaviour.
	if err := c.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("stop container %s: %w", containerID, err)
	}

	if err := c.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true, RemoveVolumes: true}); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("remove container %s: %w", containerID, err)
	}

	return nil
}

func (c *DockerClient) CommitContainer(ctx context.Context, containerID, targetImage string) error {
	if targetImage == "" {
		return fmt.Errorf("invalid target image reference")
	}

	if _, err := c.client.ContainerCommit(ctx, containerID, types.ContainerCommitOptions{
		Reference: targetImage,
		Author:    c.dockerCfg.SnapshotAuthor,
		Comment:   c.dockerCfg.SnapshotComment,
	}); err != nil {
		return fmt.Errorf("commit container %s to %s: %w", containerID, targetImage, err)
	}

	return nil
}

func (c *DockerClient) buildTraefikLabels(containerName, hostname string) map[string]string {
	// Build all dynamic router/service labels from config so every sandbox uses
	// the same Traefik conventions.
	labels := map[string]string{
		"sandbox_container": "true",
		"traefik.enable":    strconv.FormatBool(c.dockerCfg.TraefikEnable),
	}

	if c.dockerCfg.Network != "" {
		labels["traefik.docker.network"] = c.dockerCfg.Network
	}

	routerPrefix := "traefik.http.routers." + containerName
	servicePrefix := "traefik.http.services." + containerName
	labels[routerPrefix+".rule"] = fmt.Sprintf("Host(`%s`)", hostname)
	labels[servicePrefix+".loadbalancer.server.port"] = strconv.Itoa(c.sandboxCfg.InternalPort)

	if c.dockerCfg.TraefikEntrypoints != "" {
		labels[routerPrefix+".entrypoints"] = c.dockerCfg.TraefikEntrypoints
	}
	if c.dockerCfg.TraefikCertResolver != "" {
		labels[routerPrefix+".tls"] = "true"
		labels[routerPrefix+".tls.certresolver"] = c.dockerCfg.TraefikCertResolver
	}
	if c.dockerCfg.TraefikMiddlewares != "" {
		labels[routerPrefix+".middlewares"] = c.dockerCfg.TraefikMiddlewares
	}

	return labels
}
