package docker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	dockerevents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
)

type SandboxCreateRequest struct {
	ImageName     string
	RegistryRef   string
	ContainerName string
	Hostname      string
	SandboxID     string
	TTL           string
	ExpiresAt     string
	ClientIP      string
	Metadata      map[string]string
}

type SandboxContainer struct {
	ID   string
	Name string
	URL  string
	Port *int
}

type SandboxContainerEvent struct {
	ContainerID string
	Action      string
}

type Client interface {
	ImageExists(ctx context.Context, imageName string) bool
	EnsureImage(ctx context.Context, imageName string) error
	PullImage(ctx context.Context, imageName string) (io.ReadCloser, error)
	RemoveImage(ctx context.Context, imageName string) error
	CreateContainer(ctx context.Context, request SandboxCreateRequest) (*SandboxContainer, error)
	DeleteContainer(ctx context.Context, containerID string, imageName string) error
	CommitContainer(ctx context.Context, containerID, targetImage string) error
	SubscribeSandboxEvents(ctx context.Context) (<-chan SandboxContainerEvent, <-chan error)
}

type DockerClient struct {
	client     *client.Client
	sandboxCfg config.SandboxConfig
	dockerCfg  config.DockerConfig
	resolver   *registry.Resolver
	executor   *registry.Executor
}

func NewClient(sandboxCfg config.SandboxConfig, dockerCfg config.DockerConfig, resolver *registry.Resolver) (*DockerClient, error) {
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
		resolver:   resolver,
		executor:   &registry.Executor{Client: cli},
	}, nil
}

func (c *DockerClient) ImageExists(ctx context.Context, imageName string) bool {
	_, _, err := c.client.ImageInspectWithRaw(ctx, imageName)
	return err == nil
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
	reader, err := c.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("consume image pull output for %s: %w", imageName, err)
	}

	return nil
}

func (c *DockerClient) PullImage(ctx context.Context, imageName string) (io.ReadCloser, error) {
	if imageName == "" {
		return nil, fmt.Errorf("invalid image reference")
	}
	return c.client.ImagePull(ctx, imageName, image.PullOptions{})
}

func (c *DockerClient) RemoveImage(ctx context.Context, imageName string) error {
	if imageName == "" {
		return fmt.Errorf("invalid image reference")
	}

	if _, err := c.client.ImageRemove(ctx, imageName, image.RemoveOptions{Force: false, PruneChildren: false}); err != nil {
		return fmt.Errorf("remove image %s: %w", imageName, err)
	}

	return nil
}

func (c *DockerClient) CreateContainer(ctx context.Context, request SandboxCreateRequest) (*SandboxContainer, error) {
	if err := c.EnsureImage(ctx, request.ImageName); err != nil {
		return nil, err
	}

	if c.dockerCfg.Mode == config.DockerModePort {
		return c.createPortContainer(ctx, request)
	}
	return c.createTraefikContainer(ctx, request)
}

func registryName(req SandboxCreateRequest) string {
	if req.RegistryRef != "" {
		return req.RegistryRef
	}
	return req.ImageName
}

func (c *DockerClient) scheme() string {
	if c.dockerCfg.Mode == config.DockerModeTraefik && c.dockerCfg.TraefikCertResolver != "" {
		return "https"
	}
	return "http"
}

func (c *DockerClient) createPortContainer(ctx context.Context, request SandboxCreateRequest) (*SandboxContainer, error) {
	hostPort, err := findFreePort()
	if err != nil {
		return nil, fmt.Errorf("find free port: %w", err)
	}

	scheme := c.scheme()
	shopDomain := fmt.Sprintf("localhost:%d", hostPort)
	imageRepo, imageTag := splitImageRef(request.ImageName)
	resolved, err := c.resolver.Resolve(registryName(request), registry.TemplateContext{
		Hostname:       shopDomain,
		URL:            scheme + "://" + shopDomain,
		Scheme:         scheme,
		Port:           strconv.Itoa(hostPort),
		ContainerName:  request.ContainerName,
		TrustedProxies: c.dockerCfg.TrustedProxies,
		DockerMode:     string(c.dockerCfg.Mode),
		Network:        c.dockerCfg.Network,
		InternalPort:   strconv.Itoa(c.sandboxCfg.InternalPort),
		ImageName:      request.ImageName,
		ImageRepo:      imageRepo,
		ImageTag:       imageTag,
		SandboxID:      request.SandboxID,
		HostSuffix:     c.sandboxCfg.HostSuffix,
		TTL:            request.TTL,
		ExpiresAt:      request.ExpiresAt,
		ClientIP:       request.ClientIP,
		Meta:           request.Metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve registry for %s: %w", request.ImageName, err)
	}

	port := c.sandboxCfg.InternalPort
	if resolved.InternalPort > 0 {
		port = resolved.InternalPort
	}
	internalPort := nat.Port(strconv.Itoa(port) + "/tcp")

	labels := map[string]string{"sandbox_container": "true"}
	for k, v := range resolved.Labels {
		labels[k] = v
	}

	containerConfig := &container.Config{
		Image:  request.ImageName,
		Labels: labels,
		ExposedPorts: nat.PortSet{
			internalPort: struct{}{},
		},
		Env: resolved.Env,
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			internalPort: []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: strconv.Itoa(hostPort)},
			},
		},
	}

	resp, err := c.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, request.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("create container %s: %w", request.ContainerName, err)
	}

	if err := c.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("start container %s: %w", resp.ID, err)
	}

	if len(resolved.PostStart) > 0 {
		go c.executor.RunPostStart(context.Background(), resp.ID, resolved.PostStart)
	}

	return &SandboxContainer{
		ID:   resp.ID,
		Name: request.ContainerName,
		URL:  scheme + "://" + shopDomain,
		Port: &hostPort,
	}, nil
}

func splitImageRef(ref string) (string, string) {
	if strings.Contains(ref, "@") {
		return ref, ""
	}
	if i := strings.LastIndex(ref, ":"); i > strings.LastIndex(ref, "/") && i >= 0 {
		return ref[:i], ref[i+1:]
	}
	return ref, ""
}

func findFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = l.Close()
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func (c *DockerClient) createTraefikContainer(ctx context.Context, request SandboxCreateRequest) (*SandboxContainer, error) {
	scheme := c.scheme()
	imageRepo, imageTag := splitImageRef(request.ImageName)
	resolved, err := c.resolver.Resolve(registryName(request), registry.TemplateContext{
		Hostname:       request.Hostname,
		URL:            scheme + "://" + request.Hostname,
		Scheme:         scheme,
		ContainerName:  request.ContainerName,
		TrustedProxies: c.dockerCfg.TrustedProxies,
		DockerMode:     string(c.dockerCfg.Mode),
		Network:        c.dockerCfg.Network,
		InternalPort:   strconv.Itoa(c.sandboxCfg.InternalPort),
		ImageName:      request.ImageName,
		ImageRepo:      imageRepo,
		ImageTag:       imageTag,
		SandboxID:      request.SandboxID,
		HostSuffix:     c.sandboxCfg.HostSuffix,
		TTL:            request.TTL,
		ExpiresAt:      request.ExpiresAt,
		ClientIP:       request.ClientIP,
		Meta:           request.Metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve registry for %s: %w", request.ImageName, err)
	}

	port := c.sandboxCfg.InternalPort
	if resolved.InternalPort > 0 {
		port = resolved.InternalPort
	}

	labels := c.buildTraefikLabels(request.ContainerName, request.Hostname, port)
	for k, v := range resolved.Labels {
		labels[k] = v
	}

	containerConfig := &container.Config{
		Image:  request.ImageName,
		Labels: labels,
		Env:    resolved.Env,
	}

	var networkingConfig *network.NetworkingConfig
	if c.dockerCfg.Network != "" {
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

	if err := c.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("start container %s: %w", resp.ID, err)
	}

	if len(resolved.PostStart) > 0 {
		go c.executor.RunPostStart(context.Background(), resp.ID, resolved.PostStart)
	}

	return &SandboxContainer{
		ID:   resp.ID,
		Name: request.ContainerName,
		URL:  scheme + "://" + request.Hostname,
	}, nil
}

func (c *DockerClient) DeleteContainer(ctx context.Context, containerID string, imageName string) error {
	if imageName != "" {
		resolved, err := c.resolver.Resolve(imageName, registry.TemplateContext{ImageName: imageName})
		if err != nil {
			slog.Warn("pre-stop resolve failed, skipping", "container_id", containerID, "image", imageName, "error", err)
		} else if len(resolved.PreStop) > 0 {
			preStopCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			c.executor.RunPreStop(preStopCtx, containerID, resolved.PreStop)
		}
	}

	timeout := 0
	if err := c.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("stop container %s: %w", containerID, err)
	}

	if err := c.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true, RemoveVolumes: true}); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("remove container %s: %w", containerID, err)
	}

	return nil
}

func (c *DockerClient) CommitContainer(ctx context.Context, containerID, targetImage string) error {
	if targetImage == "" {
		return fmt.Errorf("invalid target image reference")
	}

	// we need to pause the source container and freeze it because ongoing mysql writes/reads makes the process crash
	if err := c.client.ContainerPause(ctx, containerID); err != nil {
		return fmt.Errorf("pause container %s before commit: %w", containerID, err)
	}
	defer func() {
		_ = c.client.ContainerUnpause(ctx, containerID)
	}()

	if _, err := c.client.ContainerCommit(ctx, containerID, container.CommitOptions{
		Reference: targetImage,
		Author:    c.dockerCfg.SnapshotAuthor,
		Comment:   c.dockerCfg.SnapshotComment,
		Pause:     true,
	}); err != nil {
		return fmt.Errorf("commit container %s to %s: %w", containerID, targetImage, err)
	}

	return nil
}

func (c *DockerClient) SubscribeSandboxEvents(ctx context.Context) (<-chan SandboxContainerEvent, <-chan error) {
	args := filters.NewArgs()
	args.Add("type", "container")
	args.Add("label", "sandbox_container=true")
	args.Add("event", "start")
	args.Add("event", "stop")
	args.Add("event", "die")
	args.Add("event", "destroy")

	msgs, errs := c.client.Events(ctx, dockerevents.ListOptions{Filters: args})

	out := make(chan SandboxContainerEvent)
	errOut := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errOut)

		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-errs:
				if !ok {
					return
				}
				if err != nil && ctx.Err() == nil {
					errOut <- err
				}
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				out <- SandboxContainerEvent{
					ContainerID: msg.ID,
					Action:      string(msg.Action),
				}
			}
		}
	}()

	return out, errOut
}

func (c *DockerClient) buildTraefikLabels(containerName, hostname string, internalPort int) map[string]string {
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
	labels[servicePrefix+".loadbalancer.server.port"] = strconv.Itoa(internalPort)

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
