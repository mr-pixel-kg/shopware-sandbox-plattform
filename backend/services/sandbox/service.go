package sandbox

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/database/models"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/database/repository"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services/images"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type SandboxService struct {
	client            *client.Client
	imageService      *images.ImageService
	sandboxRepository *repository.SandboxRepository
}

func NewSandboxService(imageService *images.ImageService, sandboxRepository *repository.SandboxRepository) (*SandboxService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	var sandboxService = &SandboxService{
		client:            cli,
		imageService:      imageService,
		sandboxRepository: sandboxRepository,
	}

	// start garbage collector scheduler
	go sandboxService.startGarbageCollectorScheduler()

	return sandboxService, nil
}

func (s *SandboxService) ListSandboxes(ctx context.Context) ([]SandboxInfo, error) {
	containers, err := s.client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}

	//fmt.Printf("Docker Containers %+v\n", containers)
	log.Printf("Found docker container %s (id: %s)\n", containers[0].Names[0], containers[0].ID)

	sandboxInfos := make([]SandboxInfo, 0)
	for _, cont := range containers {
		if cont.Labels["sandbox_container"] != "true" {
			continue
		}

		sandbox, err := s.sandboxRepository.GetByContainerID(cont.ID)
		if err != nil {
			log.Printf("Failed to fetch info for sandbox %s, because sandbox not found: %v", cont.ID, err)
			// todo error handling
		}

		if sandbox == nil {
			// Skip if no database entry exists for this container
			continue
		}

		var destroyAt *string
		if sandbox.DestroyAt != nil {
			formattedDestroyAt := sandbox.DestroyAt.Format(time.RFC3339)
			destroyAt = &formattedDestroyAt
		}

		containerInfo := SandboxInfo{
			ID:            cont.Labels["sandbox_id"],
			ContainerName: cont.Names[0],
			ContainerId:   cont.ID,
			Url:           cont.Labels["sandbox_host"],
			Image:         cont.Image,
			CreatedAt:     sandbox.CreatedAt.Format(time.RFC3339),
			DestroyAt:     destroyAt,
			State:         cont.State,
			Status:        cont.Status,
		}
		sandboxInfos = append(sandboxInfos, containerInfo)
	}

	return sandboxInfos, nil
}

func (s *SandboxService) GetSandbox(ctx context.Context, sandboxId string) (SandboxInfo, error) {

	sandbox, err := s.sandboxRepository.GetByID(sandboxId)
	if err != nil {
		log.Printf("Failed to fetch info for sandbox %s, because sandbox not found: %v", sandboxId, err)
		// todo error handling
	}
	containerId := sandbox.ContainerID

	cont, err := s.client.ContainerInspect(ctx, containerId)
	if err != nil {
		return SandboxInfo{}, err
	}

	var destroyAt *string
	if sandbox.DestroyAt != nil {
		formattedDestroyAt := sandbox.DestroyAt.Format(time.RFC3339)
		destroyAt = &formattedDestroyAt
	}

	sandboxInfo := SandboxInfo{
		ID:            cont.Config.Labels["sandbox_id"],
		ContainerName: cont.Name,
		ContainerId:   cont.ID,
		Url:           cont.Config.Labels["sandbox_host"],
		Image:         cont.Image,
		CreatedAt:     sandbox.CreatedAt.Format(time.RFC3339),
		DestroyAt:     destroyAt,
		State:         cont.State.Status,
		Status:        "up",
	}

	return sandboxInfo, nil
}

func (s *SandboxService) CreateSandbox(ctx context.Context, imageName string, lifetime int) (models.Sandbox, error) {

	sandboxId := uuid.New().String()
	containerName := "sandbox-" + sandboxId
	hostname := containerName + ".shopshredder.zion.mr-pixel.de"

	// Check if image is on whitelist
	name := strings.Split(imageName, ":")[0]
	tag := strings.Split(imageName, ":")[1]
	sandboxImage, err := s.imageService.ImageRepository.GetByNameAndTag(name, tag)
	if err != nil {
		return models.Sandbox{}, errors.New("Image " + imageName + " is not on whitelist")
	}

	// Pull docker container
	out, err := s.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		log.Print("Failed to pull sandbox docker container", err)
		return models.Sandbox{}, err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	// Create docker container
	labels := map[string]string{
		"sandbox_container": "true",
		"sandbox_id":        sandboxId,
		"sandbox_host":      hostname,
		"traefik.enable":    "true",
		fmt.Sprintf("traefik.http.routers.http-%s.rule", containerName): fmt.Sprintf("Host(`%s`)", hostname),
	}
	resp, err := s.client.ContainerCreate(ctx, &container.Config{
		Image:  imageName,
		Labels: labels,
	}, nil, nil, nil, containerName)
	if err != nil {
		log.Fatal("Failed to create sandbox docker container", err)
		return models.Sandbox{}, err
	}

	// Start docker container
	if err := s.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Print("Failed to start sandbox docker container", err)
		return models.Sandbox{}, err
	}
	log.Printf("Started sandbox %s with image %s", containerName, imageName)

	// Register sandbox in database
	var destroyAt *time.Time = nil
	if lifetime > 0 {
		tempTime := time.Now().Add(time.Minute * time.Duration(lifetime))
		destroyAt = &tempTime
	}

	sandbox := &models.Sandbox{
		ID:            sandboxId,
		ContainerID:   resp.ID,
		ContainerName: containerName,
		ImageID:       sandboxImage.ID,
		URL:           "https://" + hostname,
		CreatedAt:     time.Now(),
		DestroyAt:     destroyAt,
	}
	s.sandboxRepository.Create(sandbox)

	return *sandbox, nil
}

func (s *SandboxService) DeleteSandbox(ctx context.Context, sandboxId string) {

	// Find containerId for sandboxId
	sandbox, err := s.sandboxRepository.GetByID(sandboxId)
	if err != nil {
		log.Printf("Failed to delete sandbox %s, because sandbox not found: %v", sandboxId, err)
	}
	if sandbox == nil {
		log.Printf("Sandbox %s not found", sandboxId)
		return
	}

	// Stop sandbox container
	noWaitTimeout := 0 // to not wait for the container to exit gracefully
	err = s.client.ContainerStop(ctx, sandbox.ContainerID, container.StopOptions{Timeout: &noWaitTimeout})
	if err != nil {
		log.Printf("Failed to stop sandbox container %s: %v", sandbox.ContainerName, err)
	}

	// Delete sandbox container
	err = s.client.ContainerRemove(ctx, sandbox.ContainerID, container.RemoveOptions{Force: true, RemoveVolumes: true})
	if err != nil {
		log.Printf("Failed to remove sandbox container %s: %v", sandbox.ContainerName, err)
		return
	}

	// Remove sandbox from database
	err = s.sandboxRepository.Delete(sandboxId)
	if err != nil {
		log.Printf("Failed to delete sandbox %s in database: %v", sandboxId, err)
	}
}

func (s *SandboxService) ShutdownSandboxes() {
	ctx := context.Background()

	containers, err := s.sandboxRepository.GetExpiredContainers()
	if err != nil {
		fmt.Printf("Error getting expired containers: %v", err)
		return
	}

	for _, container := range containers {
		log.Println(container.ContainerName + " is expired. Shutting down...")
		s.DeleteSandbox(ctx, container.ID)
	}
}

type SandboxInfo struct {
	ID            string  `json:"id"`
	ContainerId   string  `json:"container_id"`
	ContainerName string  `json:"container_name"`
	Url           string  `json:"url"`
	Image         string  `json:"image"`
	CreatedAt     string  `json:"created_at"`
	DestroyAt     *string `json:"destroy_at"`
	State         string  `json:"state"`
	Status        string  `json:"status"`
}

const gcInterval = 10 * time.Minute

func (s *SandboxService) startGarbageCollectorScheduler() {
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.ShutdownSandboxes()
			s.garbageCollect()
		}
	}
}

func (s *SandboxService) garbageCollect() {
	log.Println("Check for expired sandbox containers...")
	s.ShutdownSandboxes()

	// TODO check for dangling database records which have no corresponding container
}
