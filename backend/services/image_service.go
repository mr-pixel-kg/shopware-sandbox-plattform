package services

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database/repository"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"
)

type ImageService struct {
	dockerService   *DockerService
	imageRepository *repository.ImageRepository
	client          *client.Client
}

func NewImageService(dockerService *DockerService, imageRepository *repository.ImageRepository) (*ImageService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	imageService := &ImageService{
		dockerService:   dockerService,
		imageRepository: imageRepository,
		client:          cli,
	}
	imageService.startupCheck()

	return imageService, nil
}

func (s *ImageService) ListImages(ctx context.Context) ([]Image, error) {
	// List images
	images, err := s.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	var output = make([]Image, 0)

	for _, image := range images {

		// This is a hacky fix to avoid errors by images like "nginx" without tag
		// TODO: Rewrite this filter ; IMAGE id is a hash and has no ":" !!!??? FORMAT is sha256:hash
		if strings.Contains(image.ID, ":") == false || len(image.RepoTags) < 1 || strings.Contains(image.RepoTags[0], ":") == false {
			continue
		}

		// Extract id
		imageHash := strings.Split(image.ID, ":")[1]

		// Extract image name (only uses first)
		imageName := strings.Split(image.RepoTags[0], ":")[0]

		// Extract image tag (only uses first)
		imageTag := strings.Split(image.RepoTags[0], ":")[1]

		// Check if image is on whitelist
		if !s.imageRepository.IsAllowed(imageName, imageTag) {
			continue
		}

		output = append(output, Image{
			ID:        imageHash,
			ImageName: imageName,
			ImageTag:  imageTag,
			CreatedAt: time.Unix(image.Created, 0).Format(time.RFC3339),
			Size:      image.Size,
		})
	}

	return output, nil
}

func (s *ImageService) GetImage(ctx context.Context, imageId string) (Image, error) {

	// Get image
	image, _, err := s.client.ImageInspectWithRaw(ctx, imageId)
	if err != nil {
		return Image{}, err
	}

	// Extract id
	imageHash := strings.Split(image.ID, ":")[1]

	// Extract image name (only uses first)
	imageName := strings.Split(image.RepoTags[0], ":")[0]

	// Extract image tag (only uses first)
	imageTag := getTagFromImageName(image.RepoTags[0])

	// Check if image is on whitelist
	if !s.imageRepository.IsAllowed(imageName, imageTag) {
		return Image{}, errors.New("Requested docker image is not on whitelist")
	}

	return Image{
		ID:        imageHash,
		ImageName: imageName,
		ImageTag:  imageTag,
		CreatedAt: image.Created,
		Size:      image.Size,
	}, nil
}

func (s *ImageService) PullImage(ctx context.Context, imageName string) (Image, error) {

	// Check if image already exists locally
	exists := true
	_, _, err2 := s.client.ImageInspectWithRaw(context.Background(), imageName)
	if err2 != nil {
		// Falls das Image nicht existiert, gibt Docker einen speziellen Fehler zurück
		if client.IsErrNotFound(err2) {
			exists = false
			slog.Info("Docker Image not found locally", "imageName", imageName)
		} else {
			// Sonstiger fehler
			return Image{}, err2
		}
	}

	if !exists {
		slog.Info("Pull Docker Image ...", "imageName", imageName)
		out, err := s.client.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			slog.Error("Failed to pull docker image", "imageName", imageName, "err", err)
			return Image{}, err
		}

		defer out.Close()

		io.Copy(os.Stdout, out)
	}

	image, _, err := s.client.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		slog.Error("Failed to read docker image details", "imageName", imageName, "err", err)
		return Image{}, err
	}

	// Extract id
	imageHash := strings.Split(image.ID, ":")[1]

	// Extract image name (only uses first)
	imageName = strings.Split(image.RepoTags[0], ":")[0]

	// Extract image tag (only uses first)
	imageTag := getTagFromImageName(image.RepoTags[0])

	// Add image to whitelist
	_, err = s.imageRepository.Create(imageName, imageTag)
	if err != nil {
		return Image{}, err
	}
	// todo: return image-id in response

	return Image{
		ID:        imageHash,
		ImageName: imageName,
		ImageTag:  imageTag,
		CreatedAt: image.Created,
		Size:      image.Size,
	}, nil
}

func (s *ImageService) DeleteImage(ctx context.Context, imageId string) error {
	// Get img
	img, _, err := s.client.ImageInspectWithRaw(ctx, imageId)
	if err != nil {
		return err
	}

	// Extract img name (only uses first)
	imageName := strings.Split(img.RepoTags[0], ":")[0]

	// Extract img tag (only uses first)
	imageTag := getTagFromImageName(img.RepoTags[0])

	// Check if img is on whitelist
	if !s.imageRepository.IsAllowed(imageName, imageTag) {
		return errors.New("Requested docker img is not on whitelist")
	}

	// Delete docker image
	_, err = s.client.ImageRemove(ctx, imageId, image.RemoveOptions{Force: false})
	if err != nil {
		return err
	}

	// Remove image from whitelist
	err = s.imageRepository.DeleteByTagAndName(imageName, imageTag)
	return nil
}

func (s *ImageService) startupCheck() {
	log.Println("*** Executing image service startup check ***")

	images, err := s.ListImages(context.Background())
	if err != nil {
		log.Panicf("Failed to list docker images: %v", err)
	}

	for _, img := range images {
		log.Printf("Found sandbox image: %s:%s", img.ImageName, img.ImageTag)
	}
}

type Image struct {
	ID        string `json:"id" example:"a407dee395ed97ead1e40c7537395d6271c07cc89c317f8eda1c19f6fc783695"`
	ImageName string `json:"image_name" example:"dockware/dev"`
	ImageTag  string `json:"image_tag" example:"6.6.8.2"`
	CreatedAt string `json:"created_at" example:"2013-08-20T18:52:09.000Z"`
	Size      int64  `json:"size" example:"1048576"`
}

// getTagFromImageName extracts the tag from an image name (e.g., "alpine:latest" -> "latest")
func getTagFromImageName(imageName string) string {
	parts := strings.Split(imageName, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return "latest"
}
