package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/apperror"
	auditcontracts "github.com/mr-pixel-kg/shopshredder/api/internal/auditlog"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/errs"
	mw "github.com/mr-pixel-kg/shopshredder/api/internal/http/middleware"
	"github.com/mr-pixel-kg/shopshredder/api/internal/models"
	"github.com/mr-pixel-kg/shopshredder/api/internal/registry"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
	"gorm.io/gorm"
)

type RegistryResolver interface {
	SchemaFor(imageName string) *registry.MetadataSchema
}

type ImageHandler struct {
	Images   *services.ImageService
	Audit    *services.AuditService
	Resolver RegistryResolver
}

func (h ImageHandler) ListImages(c fuego.ContextNoBody) (dto.ImageListResponse, error) {
	r := c.Request()
	auth := mw.AuthFromContext(r)

	limit, offset, err := parsePaginationParams(r)
	if err != nil {
		return dto.ImageListResponse{}, err
	}

	input := services.ImageListInput{Limit: limit, Offset: offset}

	var result *services.ImageListResult
	if auth == nil || r.URL.Query().Get("visibility") == "public" {
		result, err = h.Images.ListPublicPaginated(input)
	} else {
		result, err = h.Images.ListAllPaginated(input)
	}
	if err != nil {
		return dto.ImageListResponse{}, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not load images"}
	}

	meta := h.Images.EnrichMetadata(result.Images)
	out := make([]dto.ImageResponse, len(result.Images))
	for i := range result.Images {
		img := &result.Images[i]
		out[i] = imageToResponse(img, meta[img.ID])
	}
	return dto.ImageListResponse{
		Data: out,
		Meta: dto.PaginatedMeta{
			Pagination: buildPaginationMeta(len(out), result.Limit, result.Offset, result.Total),
		},
	}, nil
}

func (h ImageHandler) Create(c fuego.ContextWithBody[dto.CreateImageRequest]) (dto.ImageResponse, error) {
	body, err := c.Body()
	if err != nil {
		return dto.ImageResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	auth := mw.MustAuth(c.Request())
	slog.Debug("image creation requested", "component", "image", "user_id", auth.UserID, "name", body.Name, "tag", body.Tag)

	image, err := h.Images.CreateForUser(
		c.Request().Context(),
		&auth.UserID,
		body.Name, body.Tag,
		body.Title, body.Description,
		body.IsPublic,
		body.Metadata, nil,
	)
	if err != nil {
		return dto.ImageResponse{}, mapImageError(err)
	}

	resourceType := auditcontracts.ResourceTypeImage
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionImageCreated, &resourceType, &image.ID, map[string]any{
		"name": body.Name, "tag": body.Tag,
	}))

	slog.Info("image created", "component", "image", "user_id", auth.UserID, "image_id", image.ID, "image", image.FullName(), "status", image.Status)
	return imageResponseFor(h.Images, image), nil
}

func (h ImageHandler) Update(c fuego.ContextWithBody[dto.UpdateImageRequest]) (dto.ImageResponse, error) {
	auth := mw.MustAuth(c.Request())
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return dto.ImageResponse{}, err
	}

	body, err := c.Body()
	if err != nil {
		return dto.ImageResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	image, err := h.Images.Update(id, body.Title, body.Description, body.IsPublic, body.Metadata)
	if err != nil {
		return dto.ImageResponse{}, mapImageError(err)
	}

	slog.Info("image updated", "component", "image", "user_id", auth.UserID, "image_id", image.ID)
	resourceType := auditcontracts.ResourceTypeImage
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionImageUpdated, &resourceType, &image.ID, map[string]any{}))
	return imageResponseFor(h.Images, image), nil
}

func (h ImageHandler) Delete(c fuego.ContextNoBody) (any, error) {
	auth := mw.MustAuth(c.Request())
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return nil, err
	}

	slog.Debug("image deletion requested", "component", "image", "user_id", auth.UserID, "image_id", id)
	if err := h.Images.Delete(c.Request().Context(), id); err != nil {
		return nil, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not delete image"}
	}

	slog.Info("image deleted", "component", "image", "user_id", auth.UserID, "image_id", id)
	resourceType := auditcontracts.ResourceTypeImage
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionImageDeleted, &resourceType, &id, map[string]any{}))
	return nil, nil
}

func (h ImageHandler) UploadThumbnail(w http.ResponseWriter, r *http.Request) {
	auth := mw.MustAuth(r)
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		errs.Write(w, http.StatusBadRequest, "Invalid image id")
		return
	}

	file, fh, err := r.FormFile("thumbnail")
	if err != nil {
		errs.Write(w, http.StatusBadRequest, "Missing thumbnail upload")
		return
	}
	defer file.Close()

	slog.Debug("thumbnail upload requested", "component", "image", "user_id", auth.UserID, "image_id", id, "filename", fh.Filename, "size", fh.Size)
	image, err := h.Images.SaveThumbnail(id, file, fh.Filename, fh.Header.Get("Content-Type"))
	if err != nil {
		if errors.Is(err, services.ErrUnsupportedThumbnailFormat) {
			errs.Write(w, http.StatusBadRequest, "Unsupported thumbnail format")
			return
		}
		errs.Write(w, http.StatusInternalServerError, "Could not store thumbnail")
		return
	}

	slog.Info("thumbnail uploaded", "component", "image", "user_id", auth.UserID, "image_id", image.ID)
	resourceType := auditcontracts.ResourceTypeImage
	_ = h.Audit.Log(newAuditLogInput(r, &auth.UserID, auditcontracts.ActionImageThumbnailUploaded, &resourceType, &image.ID, map[string]any{}))

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(imageResponseFor(h.Images, image))
}

func (h ImageHandler) DeleteThumbnail(w http.ResponseWriter, r *http.Request) {
	auth := mw.MustAuth(r)
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		errs.Write(w, http.StatusBadRequest, "Invalid image id")
		return
	}

	image, err := h.Images.DeleteThumbnail(id)
	if err != nil {
		errs.Write(w, http.StatusInternalServerError, "Could not delete thumbnail")
		return
	}

	slog.Info("thumbnail deleted", "component", "image", "user_id", auth.UserID, "image_id", image.ID)
	resourceType := auditcontracts.ResourceTypeImage
	_ = h.Audit.Log(newAuditLogInput(r, &auth.UserID, auditcontracts.ActionImageThumbnailDeleted, &resourceType, &image.ID, map[string]any{}))
	w.WriteHeader(http.StatusNoContent)
}

func (h ImageHandler) ListPending(c fuego.ContextNoBody) ([]dto.PendingImageResponse, error) {
	images, percents := h.Images.ListPendingImages()
	out := make([]dto.PendingImageResponse, len(images))
	for i, img := range images {
		out[i] = dto.PendingImageResponse{
			ID:      img.ID,
			Name:    img.Name,
			Tag:     img.Tag,
			Title:   img.Title,
			Percent: percents[img.ID.String()],
			Status:  img.Status,
		}
	}
	return out, nil
}

func (h ImageHandler) Progress(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		errs.Write(w, http.StatusBadRequest, "Invalid image id")
		return
	}

	image, err := h.Images.FindByID(id)
	if err != nil {
		errs.Write(w, http.StatusNotFound, "Image not found")
		return
	}

	switch image.Status {
	case models.ImageStatusReady:
		writeSSEHeaders(w)
		sendSSEEvent(w, map[string]any{"percent": 100, "status": "ready"})

	case models.ImageStatusFailed:
		writeSSEHeaders(w)
		errMsg := ""
		if image.Error != nil {
			errMsg = *image.Error
		}
		sendSSEEvent(w, map[string]any{"percent": 0, "status": "failed", "error": errMsg})

	case models.ImageStatusPulling:
		if !h.Images.IsPulling(id.String()) {
			writeSSEHeaders(w)
			sendSSEEvent(w, map[string]any{"percent": 0, "status": "failed", "error": "pull process not running"})
			return
		}

		writeSSEHeaders(w)
		ch, cancel := h.Images.WatchPullProgress(id.String())
		defer cancel()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case progress, ok := <-ch:
				if !ok {
					return
				}
				sendSSEEvent(w, progress)
			}
		}

	case models.ImageStatusCommitting:
		writeSSEHeaders(w)
		sendSSEEvent(w, map[string]any{"percent": 0, "status": "committing"})

	default:
		errs.Write(w, http.StatusNotFound, "Image not found")
	}
}

func (h ImageHandler) RegistryLookup(c fuego.ContextNoBody) (registry.MetadataSchema, error) {
	r := c.Request()
	name := r.URL.Query().Get("name")

	if name == "" {
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			id, err := uuid.Parse(idStr)
			if err != nil {
				return registry.MetadataSchema{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid image id"}
			}
			image, err := h.Images.FindByID(id)
			if err != nil {
				return registry.MetadataSchema{}, fuego.HTTPError{Status: http.StatusNotFound, Detail: "Image not found"}
			}
			name = image.RegistryName()
		}
	}

	if name == "" {
		return registry.MetadataSchema{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "name or id query parameter is required"}
	}

	if schema := h.Resolver.SchemaFor(name); schema != nil {
		return *schema, nil
	}
	return registry.MetadataSchema{Items: []registry.MetadataItem{}}, nil
}

func imageResponseFor(s *services.ImageService, img *models.Image) dto.ImageResponse {
	meta := s.EnrichMetadata([]models.Image{*img})
	return imageToResponse(img, meta[img.ID])
}

func imageToResponse(img *models.Image, metadata []registry.MetadataItem) dto.ImageResponse {
	var owner *dto.UserSummary
	if img.Owner != nil {
		owner = &dto.UserSummary{ID: img.Owner.ID, Email: img.Owner.Email, AvatarURL: dto.GravatarURL(img.Owner.Email, 80)}
	}
	if metadata == nil {
		metadata = []registry.MetadataItem{}
	}
	return dto.ImageResponse{
		ID: img.ID, Name: img.Name, Tag: img.Tag,
		Title: img.Title, Description: img.Description,
		ThumbnailURL: img.ThumbnailURL, IsPublic: img.IsPublic,
		Status: img.Status, Error: img.Error,
		Metadata: metadata, RegistryRef: img.RegistryRef,
		Owner:     owner,
		CreatedAt: img.CreatedAt, UpdatedAt: img.UpdatedAt,
	}
}

func mapImageError(err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return fuego.HTTPError{Status: appErr.StatusCode, Detail: appErr.Message}
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fuego.HTTPError{Status: http.StatusNotFound, Detail: "Image not found"}
	}
	return fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Image operation failed"}
}

func writeSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
}

func sendSSEEvent(w http.ResponseWriter, v any) {
	data, _ := json.Marshal(v)
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
