package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
)

type ImageService struct {
	repo   *repositories.ImageRepository
	docker docker.Client
}

func NewImageService(repo *repositories.ImageRepository, dockerClient docker.Client) *ImageService {
	return &ImageService{
		repo:   repo,
		docker: dockerClient,
	}
}

func (s *ImageService) ListPublic() ([]models.Image, error) {
	return s.repo.ListPublic()
}

func (s *ImageService) ListAll() ([]models.Image, error) {
	return s.repo.ListAll()
}

func (s *ImageService) Create(image *models.Image) error {
	image.ID = uuid.New()
	return s.repo.Create(image)
}

func (s *ImageService) FindByID(id uuid.UUID) (*models.Image, error) {
	return s.repo.FindByID(id)
}

func (s *ImageService) CreateForUser(
	ctx context.Context,
	userID *uuid.UUID,
	name string,
	tag string,
	title *string,
	description *string,
	thumbnailURL *string,
	isPublic bool,
) (*models.Image, error) {
	fullName := name + ":" + tag
	if err := s.docker.EnsureImage(ctx, fullName); err != nil {
		return nil, err
	}

	image := &models.Image{
		ID:              uuid.New(),
		Name:            name,
		Tag:             tag,
		Title:           title,
		Description:     description,
		ThumbnailURL:    thumbnailURL,
		IsPublic:        isPublic,
		CreatedByUserID: userID,
	}

	if err := s.repo.Create(image); err != nil {
		return nil, err
	}

	return image, nil
}

func (s *ImageService) Delete(ctx context.Context, id uuid.UUID) error {
	image, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.docker.RemoveImage(ctx, image.FullName()); err != nil {
		return err
	}

	return s.repo.Delete(id)
}
