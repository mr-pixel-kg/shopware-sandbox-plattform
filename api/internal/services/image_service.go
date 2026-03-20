package services

import (
	"context"
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
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
)

type PendingPull struct {
	ID           uuid.UUID
	Name         string
	Tag          string
	Title        *string
	Description  *string
	ThumbnailURL *string
	IsPublic     bool
	UserID       *uuid.UUID
	Percent      int
	Status       string // could be redundant but usefully for future currently its most of the time just "pulling"
}

const (
	ThumbnailPublicBasePath = "/thumbnails"
)

var ErrUnsupportedThumbnailFormat = errors.New("unsupported thumbnail format")

type ImageService struct {
	repo         *repositories.ImageRepository
	sandboxRepo  *repositories.SandboxRepository
	docker       docker.Client
	tracker      *docker.PullTracker
	thumbnailDir string

	mu           sync.RWMutex
	pullCancels  map[string]context.CancelFunc
	pendingPulls map[string]*PendingPull
}

func NewImageService(
	repo *repositories.ImageRepository,
	sandboxRepo *repositories.SandboxRepository,
	dockerClient docker.Client,
	tracker *docker.PullTracker,
	thumbnailDir string,
) *ImageService {
	service := &ImageService{
		repo:         repo,
		sandboxRepo:  sandboxRepo,
		docker:       dockerClient,
		tracker:      tracker,
		thumbnailDir: thumbnailDir,
		pullCancels:  make(map[string]context.CancelFunc),
		pendingPulls: make(map[string]*PendingPull),
	}

	if err := os.MkdirAll(service.thumbnailDir, 0o755); err != nil {
		slog.Error("create thumbnail directory failed", "path", service.thumbnailDir, "cause", err.Error())
	}

	return service
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

func (s *ImageService) CreateForUser(
	ctx context.Context,
	userID *uuid.UUID,
	name string,
	tag string,
	title *string,
	description *string,
	isPublic bool,
) (*models.Image, *PendingPull, error) {
	fullName := name + ":" + tag

	if s.docker.ImageExists(ctx, fullName) {
		img := &models.Image{
			ID:              uuid.New(),
			Name:            name,
			Tag:             tag,
			Title:           title,
			Description:     description,
			IsPublic:        isPublic,
			CreatedByUserID: userID,
		}
		if err := s.repo.Create(img); err != nil {
			return nil, nil, err
		}
		return img, nil, nil
	}

	pending := &PendingPull{
		ID:          uuid.New(),
		Name:        name,
		Tag:         tag,
		Title:       title,
		Description: description,
		IsPublic:    isPublic,
		UserID:      userID,
		Status:      "pulling",
	}

	pullCtx, cancel := context.WithCancel(context.Background())
	idStr := pending.ID.String()

	s.mu.Lock()
	s.pendingPulls[idStr] = pending
	s.pullCancels[idStr] = cancel
	s.mu.Unlock()

	go s.pullImage(pullCtx, pending, fullName)

	return nil, pending, nil
}

func (s *ImageService) pullImage(ctx context.Context, pending *PendingPull, fullName string) {
	idStr := pending.ID.String()
	s.tracker.Start(idStr)

	defer func() {
		s.mu.Lock()
		delete(s.pullCancels, idStr)
		delete(s.pendingPulls, idStr)
		s.mu.Unlock()
		time.AfterFunc(10*time.Second, func() { s.tracker.Remove(idStr) })
	}()

	reader, err := s.docker.PullImage(ctx, fullName)
	if err != nil {
		if ctx.Err() != nil {
			slog.Info("image pull cancelled", "image_id", idStr, "image", fullName)
		} else {
			slog.Error("image pull failed", "image_id", idStr, "image", fullName, "error", err.Error())
		}
		s.tracker.Finish(idStr, err)
		return
	}
	defer reader.Close()

	if err := s.tracker.ConsumePullStream(idStr, reader); err != nil {
		if ctx.Err() != nil {
			slog.Info("image pull cancelled", "image_id", idStr, "image", fullName)
		} else {
			slog.Error("image pull stream failed", "image_id", idStr, "image", fullName, "error", err.Error())
		}
		s.tracker.Finish(idStr, err)
		return
	}

	img := &models.Image{
		ID:              pending.ID,
		Name:            pending.Name,
		Tag:             pending.Tag,
		Title:           pending.Title,
		Description:     pending.Description,
		ThumbnailURL:    pending.ThumbnailURL,
		IsPublic:        pending.IsPublic,
		CreatedByUserID: pending.UserID,
	}
	if err := s.repo.Create(img); err != nil {
		slog.Error("failed to persist image after pull", "image_id", idStr, "error", err.Error())
		s.tracker.Finish(idStr, err)
		return
	}

	slog.Info("image pull complete", "image_id", idStr, "image", fullName)
	s.tracker.Finish(idStr, nil)

	// todo hier kein image zurückgeben? s.attachThumbnailURL(image)
}

func (s *ImageService) WatchPullProgress(imageID string) (<-chan docker.PullProgress, func()) {
	return s.tracker.Watch(imageID)
}

func (s *ImageService) Update(id uuid.UUID, title, description *string, isPublic bool) (*models.Image, error) {
	image, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	image.Title = title
	image.Description = description
	image.IsPublic = isPublic
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
	dst, err := os.Create(targetPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

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
	_, ok := s.pendingPulls[imageID]
	s.mu.RUnlock()
	return ok
}

func (s *ImageService) ListPendingPulls() []PendingPull {
	s.mu.RLock()
	copies := make([]PendingPull, 0, len(s.pendingPulls))
	ids := make([]string, 0, len(s.pendingPulls))
	for id, p := range s.pendingPulls {
		copies = append(copies, *p)
		ids = append(ids, id)
	}
	s.mu.RUnlock()

	for i, id := range ids {
		progress := s.tracker.Progress(id)
		copies[i].Percent = progress.Percent
	}
	return copies
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
	_, isPending := s.pendingPulls[idStr]
	s.mu.RUnlock()
	if isPending {
		s.cancelPull(idStr)
		return nil
	}

	img, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Delete all sandboxes that use this image first.
	sandboxes, err := s.sandboxRepo.ListByImageID(id)
	if err != nil {
		return err
	}

	for _, sb := range sandboxes {
		if sb.Status == models.SandboxStatusStarting || sb.Status == models.SandboxStatusRunning {
			if err := s.docker.DeleteContainer(ctx, sb.ContainerID); err != nil {
				slog.Warn("failed to delete sandbox container during image deletion", "container_id", sb.ContainerID, "error", err.Error())
			}
		}
		if err := s.sandboxRepo.DeleteByID(sb.ID); err != nil {
			slog.Warn("failed to delete sandbox during image deletion", "sandbox_id", sb.ID.String(), "error", err.Error())
		}
	}

	// Remove Docker image if it exists locally.
	if s.docker.ImageExists(ctx, img.FullName()) {
		if err := s.docker.RemoveImage(ctx, img.FullName()); err != nil {
			slog.Warn("docker image removal failed, proceeding with db deletion", "image", img.FullName(), "error", err.Error())
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
		slog.Error("resolve thumbnail path failed", "image_id", image.ID.String(), "cause", err.Error())
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
