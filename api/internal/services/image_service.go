package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/apperror"
	"github.com/mr-pixel-kg/shopshredder/api/internal/docker"
	"github.com/mr-pixel-kg/shopshredder/api/internal/models"
	"github.com/mr-pixel-kg/shopshredder/api/internal/registry"
	"github.com/mr-pixel-kg/shopshredder/api/internal/repositories"
	"gorm.io/datatypes"
)

const (
	ThumbnailPublicBasePath = "/thumbnails"
)

var ErrUnsupportedThumbnailFormat = errors.New("unsupported thumbnail format")

type ImageService struct {
	repo          *repositories.ImageRepository
	sandboxes     *SandboxService
	docker        docker.Client
	tracker       *docker.PullTracker
	thumbnailDir  string
	publicBaseURL string
	resolver      *registry.Resolver

	mu          sync.RWMutex
	pullCancels map[string]context.CancelFunc
}

func (s *ImageService) SetSandboxService(sandboxes *SandboxService) {
	s.sandboxes = sandboxes
}

func NewImageService(
	repo *repositories.ImageRepository,
	dockerClient docker.Client,
	tracker *docker.PullTracker,
	publicBaseURL string,
	thumbnailDir string,
	resolver *registry.Resolver,
) *ImageService {
	service := &ImageService{
		repo:          repo,
		docker:        dockerClient,
		tracker:       tracker,
		thumbnailDir:  thumbnailDir,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		resolver:      resolver,
		pullCancels:   make(map[string]context.CancelFunc),
	}

	if err := os.MkdirAll(service.thumbnailDir, 0o750); err != nil {
		slog.Error("create thumbnail directory failed", "component", "image", "path", service.thumbnailDir, "error", err.Error())
	}

	return service
}

func (s *ImageService) ReconcileOnStartup(ctx context.Context) {
	stale, err := s.repo.ListByStatuses([]string{models.ImageStatusPulling, models.ImageStatusCommitting})
	if err != nil {
		slog.Error("reconcile: failed to list pending images", "component", "image", "error", err.Error())
	}
	for _, img := range stale {
		errMsg := "operation interrupted by API restart"
		if err := s.repo.UpdateStatus(img.ID, models.ImageStatusFailed, &errMsg); err != nil {
			slog.Error("reconcile: failed to mark image as failed", "component", "image", "image_id", img.ID.String(), "error", err.Error())
		} else {
			slog.Info("reconcile: marked stale image as failed", "component", "image", "image_id", img.ID.String(), "image", img.FullName(), "previous_status", img.Status)
		}
	}

	ready, err := s.repo.ListByStatus(models.ImageStatusReady)
	if err != nil {
		slog.Error("reconcile: failed to list ready images", "component", "image", "error", err.Error())
		return
	}
	for _, img := range ready {
		if !s.docker.ImageExists(ctx, img.FullName()) {
			slog.Info("reconcile: ready image missing from docker, re-pulling", "component", "image", "image_id", img.ID.String(), "image", img.FullName())
			if err := s.repo.UpdateStatus(img.ID, models.ImageStatusPulling, nil); err != nil {
				slog.Error("reconcile: failed to update status to pulling", "component", "image", "image_id", img.ID.String(), "error", err.Error())
				continue
			}
			s.startPull(img.ID, img.FullName())
		}
	}
}

type ImageListInput struct {
	Limit  int
	Offset int
}

type ImageListResult struct {
	Images []models.Image
	Total  int64
	Limit  int
	Offset int
}

func (s *ImageService) ListAllPaginated(input ImageListInput) (*ImageListResult, error) {
	images, total, err := s.repo.ListAllPaginated(input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}
	images = s.attachThumbnailURLs(images)
	return &ImageListResult{
		Images: images,
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	}, nil
}

func (s *ImageService) ListPublicPaginated(input ImageListInput) (*ImageListResult, error) {
	images, total, err := s.repo.ListPublicPaginated(input.Limit, input.Offset)
	if err != nil {
		return nil, err
	}
	images = s.attachThumbnailURLs(images)
	return &ImageListResult{
		Images: images,
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	}, nil
}

func (s *ImageService) ListPublic() ([]models.Image, error) {
	images, err := s.repo.ListPublic()
	if err != nil {
		return nil, err
	}
	return s.attachThumbnailURLs(images), nil
}

func (s *ImageService) ListAll() ([]models.Image, error) {
	images, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	return s.attachThumbnailURLs(images), nil
}

func (s *ImageService) FindByID(id uuid.UUID) (*models.Image, error) {
	image, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.attachThumbnailURL(image), nil
}

func (s *ImageService) FindByIDs(ids []uuid.UUID) ([]models.Image, error) {
	return s.repo.FindByIDs(ids)
}

func (s *ImageService) createImage(userID *uuid.UUID, name, tag string, title, description *string, isPublic bool, metadata []registry.MetadataItem, registryRef *string, status string) (*models.Image, error) {
	img := &models.Image{
		ID:          uuid.New(),
		Name:        name,
		Tag:         tag,
		Title:       title,
		Description: description,
		IsPublic:    isPublic,
		Status:      status,
		OwnerID:     userID,
		RegistryRef: registryRef,
	}
	encoded, err := s.encodeMetadata(img, metadata)
	if err != nil {
		return nil, err
	}
	img.Metadata = encoded

	if err := s.repo.Create(img); err != nil {
		return nil, err
	}

	return s.attachThumbnailURL(img), nil
}

func (s *ImageService) encodeMetadata(img *models.Image, metadata []registry.MetadataItem) (datatypes.JSON, error) {
	schema := s.resolver.SchemaFor(img.RegistryName())
	if err := registry.ValidateMetadata(registry.MergeWithRegistry(schema, metadata), schema); err != nil {
		return nil, apperror.BadRequest("INVALID_METADATA", err.Error())
	}
	return json.Marshal(registry.StripRegistryDuplicates(metadata, schema))
}

func decodeMetadata(raw datatypes.JSON) []registry.MetadataItem {
	var metadata []registry.MetadataItem
	_ = json.Unmarshal(raw, &metadata)
	return metadata
}

func (s *ImageService) CreateForUser(
	ctx context.Context,
	userID *uuid.UUID,
	name, tag string,
	title, description *string,
	isPublic bool,
	metadata []registry.MetadataItem,
	registryRef *string,
) (*models.Image, error) {
	img, err := s.createImage(userID, name, tag, title, description, isPublic, metadata, registryRef, models.ImageStatusPulling)
	if err != nil {
		return nil, err
	}

	fullName := name + ":" + tag
	if s.docker.ImageExists(ctx, fullName) {
		img.Status = models.ImageStatusReady
		if err := s.repo.UpdateStatus(img.ID, models.ImageStatusReady, nil); err != nil {
			slog.Error("failed to mark image as ready", "component", "image", "image_id", img.ID.String(), "error", err.Error())
		}
		return img, nil
	}

	s.startPull(img.ID, fullName)
	return img, nil
}

func (s *ImageService) startPull(imageID uuid.UUID, fullName string) {
	pullCtx, cancel := context.WithCancel(context.Background())
	idStr := imageID.String()

	s.mu.Lock()
	s.pullCancels[idStr] = cancel
	s.mu.Unlock()

	go s.pullImage(pullCtx, imageID, fullName)
}

func (s *ImageService) pullImage(ctx context.Context, imageID uuid.UUID, fullName string) {
	idStr := imageID.String()
	s.tracker.Start(idStr)

	defer func() {
		s.mu.Lock()
		delete(s.pullCancels, idStr)
		s.mu.Unlock()
		time.AfterFunc(10*time.Second, func() { s.tracker.Remove(idStr) })
	}()

	reader, err := s.docker.PullImage(ctx, fullName)
	if err != nil {
		s.finishPull(idStr, imageID, fullName, err, ctx.Err() != nil)
		return
	}
	defer reader.Close()

	if err := s.tracker.ConsumePullStream(idStr, reader); err != nil {
		s.finishPull(idStr, imageID, fullName, err, ctx.Err() != nil)
		return
	}

	if err := s.repo.UpdateStatus(imageID, models.ImageStatusReady, nil); err != nil {
		slog.Error("failed to mark image as ready after pull", "component", "image", "image_id", idStr, "error", err.Error())
		s.tracker.Finish(idStr, err)
		return
	}

	slog.Info("image pull complete", "component", "image", "image_id", idStr, "image", fullName)
	s.tracker.Finish(idStr, nil)
}

func (s *ImageService) finishPull(idStr string, imageID uuid.UUID, fullName string, err error, cancelled bool) {
	if cancelled {
		slog.Info("image pull cancelled", "component", "image", "image_id", idStr, "image", fullName)
	} else {
		slog.Error("image pull failed", "component", "image", "image_id", idStr, "image", fullName, "error", err.Error())
	}

	errMsg := err.Error()
	if dbErr := s.repo.UpdateStatus(imageID, models.ImageStatusFailed, &errMsg); dbErr != nil {
		slog.Error("failed to mark image as failed", "component", "image", "image_id", idStr, "error", dbErr.Error())
	}

	s.tracker.Finish(idStr, err)
}

func (s *ImageService) WatchPullProgress(imageID string) (<-chan docker.PullProgress, func()) {
	return s.tracker.Watch(imageID)
}

func (s *ImageService) Update(id uuid.UUID, title, description *string, isPublic bool, metadata []registry.MetadataItem) (*models.Image, error) {
	image, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	image.Title = title
	image.Description = description
	image.IsPublic = isPublic
	if metadata != nil {
		encoded, err := s.encodeMetadata(image, metadata)
		if err != nil {
			return nil, err
		}
		image.Metadata = encoded
	}
	if err := s.repo.Update(image); err != nil {
		return nil, err
	}

	return s.attachThumbnailURL(image), nil
}

func (s *ImageService) SaveThumbnail(id uuid.UUID, file multipart.File, originalFilename, contentType string) (*models.Image, error) {
	image, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	ext, err := thumbnailExtension(file, originalFilename, contentType)
	if err != nil {
		return nil, err
	}

	if err := s.deleteThumbnailFiles(id); err != nil {
		return nil, err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	targetPath := filepath.Join(s.thumbnailDir, id.String()+ext)
	// #nosec G304 -- targetPath is derived from a service-owned directory and a UUID-based filename.
	dst, err := os.Create(targetPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = dst.Close()
	}()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	return s.attachThumbnailURL(image), nil
}

func (s *ImageService) DeleteThumbnail(id uuid.UUID) (*models.Image, error) {
	image, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.deleteThumbnailFiles(id); err != nil {
		return nil, err
	}

	return s.attachThumbnailURL(image), nil
}

func (s *ImageService) IsPulling(imageID string) bool {
	s.mu.RLock()
	_, ok := s.pullCancels[imageID]
	s.mu.RUnlock()
	return ok
}

func (s *ImageService) ListPendingImages() ([]models.Image, map[string]int) {
	images, err := s.repo.ListByStatuses([]string{models.ImageStatusPulling, models.ImageStatusCommitting})
	if err != nil {
		slog.Error("failed to list pending images", "component", "image", "error", err.Error())
		return nil, nil
	}

	percents := make(map[string]int, len(images))
	for _, img := range images {
		if img.Status == models.ImageStatusPulling {
			progress := s.tracker.Progress(img.ID.String())
			percents[img.ID.String()] = progress.Percent
		}
	}

	return s.attachThumbnailURLs(images), percents
}

func (s *ImageService) CreateForCommit(
	userID *uuid.UUID,
	name, tag string,
	title, description *string,
	isPublic bool,
	metadata []registry.MetadataItem,
	registryRef *string,
) (*models.Image, error) {
	return s.createImage(userID, name, tag, title, description, isPublic, metadata, registryRef, models.ImageStatusCommitting)
}

func (s *ImageService) FinishCommit(imageID uuid.UUID, commitErr error) {
	idStr := imageID.String()
	if commitErr != nil {
		errMsg := commitErr.Error()
		if err := s.repo.UpdateStatus(imageID, models.ImageStatusFailed, &errMsg); err != nil {
			slog.Error("failed to update image status", "component", "image", "image_id", idStr, "error", err.Error())
		}
		slog.Error("image commit failed", "component", "image", "image_id", idStr, "error", commitErr.Error())
		return
	}
	if err := s.repo.UpdateStatus(imageID, models.ImageStatusReady, nil); err != nil {
		slog.Error("failed to update image status", "component", "image", "image_id", idStr, "error", err.Error())
		return
	}
	slog.Info("image commit complete", "component", "image", "image_id", idStr)
}

func (s *ImageService) cancelPull(imageID string) {
	s.mu.RLock()
	cancel, ok := s.pullCancels[imageID]
	s.mu.RUnlock()
	if ok {
		cancel()
	}
}

func (s *ImageService) Delete(ctx context.Context, id uuid.UUID) error {
	idStr := id.String()

	s.mu.RLock()
	_, isPulling := s.pullCancels[idStr]
	s.mu.RUnlock()
	if isPulling {
		s.cancelPull(idStr)
	}

	img, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.sandboxes.StopActiveForImage(ctx, id); err != nil {
		return err
	}

	if s.docker.ImageExists(ctx, img.FullName()) {
		if err := s.docker.RemoveImage(ctx, img.FullName()); err != nil {
			slog.Warn("docker image removal failed, proceeding with db deletion", "component", "image", "image", img.FullName(), "error", err.Error())
		}
	}

	if err := s.deleteThumbnailFiles(id); err != nil {
		return err
	}

	return s.repo.Delete(id)
}

func (s *ImageService) attachThumbnailURLs(images []models.Image) []models.Image {
	for i := range images {
		s.attachThumbnailURL(&images[i])
	}
	return images
}

func (s *ImageService) attachThumbnailURL(image *models.Image) *models.Image {
	path, err := s.thumbnailPath(image.ID)
	if err != nil {
		slog.Error("resolve thumbnail path failed", "component", "image", "image_id", image.ID.String(), "error", err.Error())
		image.ThumbnailURL = nil
		return image
	}
	if path == "" {
		image.ThumbnailURL = nil
		return image
	}

	url := ThumbnailPublicBasePath + "/" + filepath.Base(path)
	image.ThumbnailURL = &url
	return image
}

func (s *ImageService) thumbnailPath(id uuid.UUID) (string, error) {
	matches, err := filepath.Glob(filepath.Join(s.thumbnailDir, id.String()+".*"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", nil
	}
	return matches[0], nil
}

func (s *ImageService) deleteThumbnailFiles(id uuid.UUID) error {
	matches, err := filepath.Glob(filepath.Join(s.thumbnailDir, id.String()+".*"))
	if err != nil {
		return err
	}

	for _, match := range matches {
		if err := os.Remove(match); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}

func thumbnailExtension(file multipart.File, originalFilename, contentType string) (string, error) {
	buffer := make([]byte, 512)
	readBytes, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	detectedType := http.DetectContentType(buffer[:readBytes])
	if ext := extensionForContentType(detectedType); ext != "" {
		return ext, nil
	}
	if ext := extensionForContentType(contentType); ext != "" {
		return ext, nil
	}

	switch strings.ToLower(filepath.Ext(originalFilename)) {
	case ".jpg", ".jpeg":
		return ".jpg", nil
	case ".png":
		return ".png", nil
	case ".gif":
		return ".gif", nil
	case ".webp":
		return ".webp", nil
	default:
		return "", ErrUnsupportedThumbnailFormat
	}
}

func extensionForContentType(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func (s *ImageService) EnrichMetadata(images []models.Image) map[uuid.UUID][]registry.MetadataItem {
	out := make(map[uuid.UUID][]registry.MetadataItem, len(images))
	for i := range images {
		img := &images[i]
		out[img.ID] = s.resolver.MergeMetadata(img.RegistryName(), decodeMetadata(img.Metadata))
	}
	return out
}
