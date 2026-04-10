package docker

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	dockerevents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"
	"github.com/mr-pixel-kg/shopshredder/api/internal/config"
)

type ContainerCreateRequest struct {
	ImageName     string
	ContainerName string
	Hostname      string
	Env           []string
	Labels        map[string]string
	InternalPort  int
	SSHPort       int
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
	ImageLabels(ctx context.Context, imageName string) (map[string]string, error)
	EnsureImage(ctx context.Context, imageName string) error
	PullImage(ctx context.Context, imageName string) (io.ReadCloser, error)
	RemoveImage(ctx context.Context, imageName string) error
	CreateContainer(ctx context.Context, request ContainerCreateRequest) (*SandboxContainer, error)
	DeleteContainer(ctx context.Context, containerID string) error
	CommitContainer(ctx context.Context, containerID, targetImage string) error
	ContainerExists(ctx context.Context, containerID string) bool
	ListSandboxContainerIDs(ctx context.Context) (map[string]struct{}, error)
	SubscribeSandboxEvents(ctx context.Context) (<-chan SandboxContainerEvent, <-chan error)
	CreateExecSession(ctx context.Context, containerID string, opts ExecAttachOptions) (*ExecSession, error)
	ContainerLogs(ctx context.Context, containerID string) (io.ReadCloser, error)
	ExecFollow(ctx context.Context, containerID string, cmd []string) (io.ReadCloser, error)
}

type DockerClient struct {
	client     *client.Client
	sandboxCfg config.SandboxConfig
	dockerCfg  config.DockerConfig
}

func NewClient(sdkClient *client.Client, sandboxCfg config.SandboxConfig, dockerCfg config.DockerConfig) *DockerClient {
	return &DockerClient{
		client:     sdkClient,
		sandboxCfg: sandboxCfg,
		dockerCfg:  dockerCfg,
	}
}

func (c *DockerClient) ImageExists(ctx context.Context, imageName string) bool {
	_, _, err := c.client.ImageInspectWithRaw(ctx, imageName)
	return err == nil
}

func (c *DockerClient) ContainerExists(ctx context.Context, containerID string) bool {
	_, err := c.client.ContainerInspect(ctx, containerID)
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

func (c *DockerClient) CreateContainer(ctx context.Context, request ContainerCreateRequest) (*SandboxContainer, error) {
	if c.dockerCfg.Mode == config.DockerModePort {
		return c.createPortContainer(ctx, request)
	}
	return c.createTraefikContainer(ctx, request)
}

func (c *DockerClient) createPortContainer(ctx context.Context, request ContainerCreateRequest) (*SandboxContainer, error) {
	containerConfig := &container.Config{
		Image:     request.ImageName,
		Labels:    request.Labels,
		Env:       request.Env,
		Tty:       true,
		OpenStdin: true,
	}

	// extract host port from the host (format: "localhost:prt")
	_, hostPortStr, _ := net.SplitHostPort(request.Hostname)
	hostPort, _ := strconv.Atoi(hostPortStr)

	hostConfig := &container.HostConfig{}

	if request.InternalPort > 0 {
		internalPort := nat.Port(strconv.Itoa(request.InternalPort) + "/tcp")
		containerConfig.ExposedPorts = nat.PortSet{
			internalPort: struct{}{},
		}
		hostConfig.PortBindings = nat.PortMap{
			internalPort: []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: strconv.Itoa(hostPort)},
			},
		}
	}

	if request.SSHPort > 0 {
		sshPort := nat.Port(strconv.Itoa(request.SSHPort) + "/tcp")
		if containerConfig.ExposedPorts == nil {
			containerConfig.ExposedPorts = nat.PortSet{}
		}
		containerConfig.ExposedPorts[sshPort] = struct{}{}
		if hostConfig.PortBindings == nil {
			hostConfig.PortBindings = nat.PortMap{}
		}
		hostConfig.PortBindings[sshPort] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: "0"},
		}
	}

	resp, err := c.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, request.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("create container %s: %w", request.ContainerName, err)
	}

	if err := c.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("start container %s: %w", resp.ID, err)
	}

	var url string
	if request.Hostname != "" {
		url = c.scheme() + "://" + request.Hostname
	}
	return &SandboxContainer{
		ID:   resp.ID,
		Name: request.ContainerName,
		URL:  url,
		Port: &hostPort,
	}, nil
}

func (c *DockerClient) createTraefikContainer(ctx context.Context, request ContainerCreateRequest) (*SandboxContainer, error) {
	containerConfig := &container.Config{
		Image:     request.ImageName,
		Labels:    request.Labels,
		Env:       request.Env,
		Tty:       true,
		OpenStdin: true,
	}

	var networkingConfig *network.NetworkingConfig
	if c.dockerCfg.Network != "" {
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				c.dockerCfg.Network: {},
			},
		}
	}

	var hostConfig *container.HostConfig
	if request.SSHPort > 0 {
		sshPort := nat.Port(strconv.Itoa(request.SSHPort) + "/tcp")
		containerConfig.ExposedPorts = nat.PortSet{sshPort: struct{}{}}
		hostConfig = &container.HostConfig{
			PortBindings: nat.PortMap{
				sshPort: []nat.PortBinding{
					{HostIP: "0.0.0.0", HostPort: "0"},
				},
			},
		}
	}

	resp, err := c.client.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, request.ContainerName)
	if err != nil {
		return nil, fmt.Errorf("create container %s: %w", request.ContainerName, err)
	}

	if err := c.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("start container %s: %w", resp.ID, err)
	}

	var url string
	if request.Hostname != "" {
		url = c.scheme() + "://" + request.Hostname
	}
	return &SandboxContainer{
		ID:   resp.ID,
		Name: request.ContainerName,
		URL:  url,
	}, nil
}

func (c *DockerClient) DeleteContainer(ctx context.Context, containerID string) error {
	timeout := 0
	if err := c.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("stop container %s: %w", containerID, err)
	}

	if err := c.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true, RemoveVolumes: true}); err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("remove container %s: %w", containerID, err)
	}

	return nil
}

func isEphemeralLabel(key string) bool {
	return strings.HasPrefix(key, "traefik.") || strings.HasPrefix(key, "sandbox_")
}

func (c *DockerClient) CommitContainer(ctx context.Context, containerID, targetImage string) error {
	if targetImage == "" {
		return fmt.Errorf("invalid target image reference")
	}

	inspect, err := c.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return fmt.Errorf("inspect container %s: %w", containerID, err)
	}
	if inspect.Config == nil {
		return fmt.Errorf("inspect container %s: missing config", containerID)
	}

	cleanConfig := *inspect.Config
	cleanConfig.Labels = make(map[string]string, len(inspect.Config.Labels))
	for k, v := range inspect.Config.Labels {
		if isEphemeralLabel(k) {
			continue
		}
		cleanConfig.Labels[k] = v
	}

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
		Pause:     false,
		Config:    &cleanConfig,
	}); err != nil {
		return fmt.Errorf("commit container %s to %s: %w", containerID, targetImage, err)
	}

	return nil
}

func (c *DockerClient) ImageLabels(ctx context.Context, imageName string) (map[string]string, error) {
	img, _, err := c.client.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return nil, fmt.Errorf("inspect image %s: %w", imageName, err)
	}
	return img.Config.Labels, nil
}

func (c *DockerClient) ListSandboxContainerIDs(ctx context.Context) (map[string]struct{}, error) {
	args := filters.NewArgs()
	args.Add("label", "sandbox_container=true")
	containers, err := c.client.ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil {
		return nil, fmt.Errorf("list sandbox containers: %w", err)
	}
	ids := make(map[string]struct{}, len(containers))
	for _, c := range containers {
		ids[c.ID] = struct{}{}
	}
	return ids, nil
}

func (c *DockerClient) SubscribeSandboxEvents(ctx context.Context) (<-chan SandboxContainerEvent, <-chan error) {
	args := filters.NewArgs()
	args.Add("type", "container")
	args.Add("label", "sandbox_container=true")
	args.Add("event", "start")
	args.Add("event", "stop")
	args.Add("event", "die")
	args.Add("event", "destroy")
	args.Add("event", "pause")
	args.Add("event", "unpause")

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

func (c *DockerClient) scheme() string {
	if c.dockerCfg.Mode == config.DockerModeTraefik && c.dockerCfg.TraefikCertResolver != "" {
		return "https"
	}
	return "http"
}

// FindFreePort finds an available tcp port
func FindFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = l.Close()
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// BuildTraefikLabels builds the treafik routing labels
func BuildTraefikLabels(containerName, hostname string, internalPort int, dockerCfg config.DockerConfig) map[string]string {
	labels := map[string]string{
		"sandbox_container": "true",
		"traefik.enable":    strconv.FormatBool(dockerCfg.TraefikEnable),
	}

	if dockerCfg.Network != "" {
		labels["traefik.docker.network"] = dockerCfg.Network
	}

	routerPrefix := "traefik.http.routers." + containerName
	servicePrefix := "traefik.http.services." + containerName
	labels[routerPrefix+".rule"] = fmt.Sprintf("Host(`%s`)", hostname)
	labels[routerPrefix+".service"] = containerName
	labels[servicePrefix+".loadbalancer.server.port"] = strconv.Itoa(internalPort)

	if dockerCfg.TraefikEntrypoints != "" {
		labels[routerPrefix+".entrypoints"] = dockerCfg.TraefikEntrypoints
	}
	if dockerCfg.TraefikCertResolver != "" {
		labels[routerPrefix+".tls"] = "true"
		labels[routerPrefix+".tls.certresolver"] = dockerCfg.TraefikCertResolver
	}
	if dockerCfg.TraefikMiddlewares != "" {
		labels[routerPrefix+".middlewares"] = dockerCfg.TraefikMiddlewares
	}

	return labels
}
