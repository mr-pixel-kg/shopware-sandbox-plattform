package images

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/database/repository"
	"io"
	"os"
	"strings"
	"time"
)

type ImageService struct {
	ImageRepository *repository.ImageRepository
	client          *client.Client
}

func NewImageService(imageRepository *repository.ImageRepository) (*ImageService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &ImageService{
		ImageRepository: imageRepository,
		client:          cli,
	}, nil
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
		// TODO: Rewrite this filter
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
		if !s.ImageRepository.IsAllowed(imageName, imageTag) {
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
	if !s.ImageRepository.IsAllowed(imageName, imageTag) {
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

	out, err := s.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return Image{}, err
	}

	defer out.Close()

	io.Copy(os.Stdout, out)

	image, _, err := s.client.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return Image{}, err
	}

	// Extract id
	imageHash := strings.Split(image.ID, ":")[1]

	// Extract image name (only uses first)
	imageName = strings.Split(image.RepoTags[0], ":")[0]

	// Extract image tag (only uses first)
	imageTag := getTagFromImageName(image.RepoTags[0])

	// Add image to whitelist
	_, err = s.ImageRepository.Create(imageName, imageTag)
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
	if !s.ImageRepository.IsAllowed(imageName, imageTag) {
		return errors.New("Requested docker img is not on whitelist")
	}

	// Delete docker image
	_, err = s.client.ImageRemove(ctx, imageId, image.RemoveOptions{Force: false})
	if err != nil {
		return err
	}

	// Remove image from whitelist
	err = s.ImageRepository.DeleteByTagAndName(imageName, imageTag)
	return nil
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
