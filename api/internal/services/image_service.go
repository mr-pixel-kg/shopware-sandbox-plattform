package services

import (
	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
)

type ImageService struct {
	repo *repositories.ImageRepository
}

func NewImageService(repo *repositories.ImageRepository) *ImageService {
	return &ImageService{repo: repo}
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

func (s *ImageService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *ImageService) FindByID(id uuid.UUID) (*models.Image, error) {
	return s.repo.FindByID(id)
}

func (s *ImageService) CreateForUser(
	userID *uuid.UUID,
	name string,
	tag string,
	title *string,
	description *string,
	thumbnailURL *string,
	isPublic bool,
) (*models.Image, error) {
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
