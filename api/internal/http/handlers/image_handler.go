package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	auditcontracts "github.com/manuel/shopware-testenv-platform/api/internal/auditlog"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/errs"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"gorm.io/gorm"
)

type RegistryResolver interface {
	ResolveEntry(imageName string) *registry.ImageEntry
}

type ImageHandler struct {
	Images   *services.ImageService
	Audit    *services.AuditService
	Resolver RegistryResolver
}

func (h ImageHandler) MountPublicRoutes(s *fuego.Server) {
	images := fuego.Group(s, "/images")
	fuego.Get(images, "/public", h.listPublic,
		option.Summary("List public images"),
		option.Description("Returns all images marked as public (no auth required)"),
		option.Tags("Images"),
	)
	fuego.Get(images, "/pending", h.listPending,
		option.Summary("List pending image operations"),
		option.Description("Returns all images with ongoing operations with optional progress percentage"),
		option.Tags("Images"),
	)
	fuego.GetStd(images, "/{id}/progress", h.progress,
		option.Summary("Stream image progress"),
		option.Description("SSE endpoint streaming progress events for image operations"),
		option.Tags("Images"),
	)
	fuego.Get(s, "/registry", h.registryLookup,
		option.Summary("Lookup registry metadata"),
		option.Description("Return registry-defined metadata for an image by name or ID"),
		option.Tags("Images"),
		option.Query("name", "Image name (e.g. dockware/dev:6.6.9.0)"),
		option.Query("id", "Image ID"),
	)
}

func (h ImageHandler) MountAuthedRoutes(s *fuego.Server) {
	images := fuego.Group(s, "/images")
	fuego.Get(images, "", h.listAll,
		option.Summary("List all images"),
		option.Description("Returns all images including private ones"),
		option.Tags("Images"),
	)
	fuego.Post(images, "", h.create,
		option.Summary("Create an image"),
		option.Description("Register a new Docker image. If not available locally, a background pull is started."),
		option.Tags("Images"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Patch(images, "/{id}", h.update,
		option.Summary("Update an image"),
		option.Description("Update image metadata and visibility"),
		option.Tags("Images"),
	)
	fuego.Delete(images, "/{id}", h.delete,
		option.Summary("Delete an image"),
		option.Description("Remove a Docker image registration"),
		option.Tags("Images"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
	fuego.PostStd(images, "/{id}/thumbnail", h.uploadThumbnail,
		option.Summary("Upload an image thumbnail"),
		option.Description("Upload or replace the thumbnail for an image"),
		option.Tags("Images"),
	)
	fuego.DeleteStd(images, "/{id}/thumbnail", h.deleteThumbnail,
		option.Summary("Delete an image thumbnail"),
		option.Description("Remove the thumbnail associated with an image"),
		option.Tags("Images"),
	)
}

func (h ImageHandler) listPublic(c fuego.ContextNoBody) ([]dto.ImageResponse, error) {
	images, err := h.Images.ListPublic()
	if err != nil {
		return nil, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not load public images"}
	}

	out := make([]dto.ImageResponse, len(images))
	for i := range images {
		out[i] = imageToResponse(&images[i])
	}
	return out, nil
}

func (h ImageHandler) listAll(c fuego.ContextNoBody) ([]dto.ImageResponse, error) {
	images, err := h.Images.ListAll()
	if err != nil {
		return nil, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not load images"}
	}

	out := make([]dto.ImageResponse, len(images))
	for i := range images {
		out[i] = imageToResponse(&images[i])
	}
	return out, nil
}

func (h ImageHandler) create(c fuego.ContextWithBody[dto.CreateImageRequest]) (dto.ImageResponse, error) {
	body, err := c.Body()
	if err != nil {
		return dto.ImageResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	auth := mw.MustAuth(c.Request())
	slog.Debug("image creation requested", "component", "image", "user_id", auth.UserID, "name", body.Name, "tag", body.Tag)

	metadataJSON, _ := json.Marshal(body.Metadata)
	image, err := h.Images.CreateForUser(
		c.Request().Context(),
		&auth.UserID,
		body.Name, body.Tag,
		body.Title, body.Description,
		body.IsPublic,
		metadataJSON, nil,
	)
	if err != nil {
		return dto.ImageResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: err.Error()}
	}

	resourceType := auditcontracts.ResourceTypeImage
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionImageCreated, &resourceType, &image.ID, map[string]any{
		"name": body.Name, "tag": body.Tag,
	}))

	slog.Info("image created", "component", "image", "user_id", auth.UserID, "image_id", image.ID, "image", image.FullName(), "status", image.Status)
	return imageToResponse(image), nil
}

func (h ImageHandler) update(c fuego.ContextWithBody[dto.UpdateImageRequest]) (dto.ImageResponse, error) {
	auth := mw.MustAuth(c.Request())
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return dto.ImageResponse{}, err
	}

	body, err := c.Body()
	if err != nil {
		return dto.ImageResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	metadataJSON, _ := json.Marshal(body.Metadata)
	image, err := h.Images.Update(id, body.Title, body.Description, body.IsPublic, metadataJSON)
	if err != nil {
		return dto.ImageResponse{}, mapImageError(err)
	}

	slog.Info("image updated", "component", "image", "user_id", auth.UserID, "image_id", image.ID)
	resourceType := auditcontracts.ResourceTypeImage
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionImageUpdated, &resourceType, &image.ID, map[string]any{}))
	return imageToResponse(image), nil
}

func (h ImageHandler) delete(c fuego.ContextNoBody) (any, error) {
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

func (h ImageHandler) uploadThumbnail(w http.ResponseWriter, r *http.Request) {
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
	_ = json.NewEncoder(w).Encode(imageToResponse(image))
}

func (h ImageHandler) deleteThumbnail(w http.ResponseWriter, r *http.Request) {
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

func (h ImageHandler) listPending(c fuego.ContextNoBody) ([]dto.PendingImageResponse, error) {
	images, percents := h.Images.ListPendingImages()
	out := make([]dto.PendingImageResponse, len(images))
	for i, img := range images {
		out[i] = dto.PendingImageResponse{
			ID:      img.ID.String(),
			Name:    img.Name,
			Tag:     img.Tag,
			Title:   img.Title,
			Percent: percents[img.ID.String()],
			Status:  img.Status,
		}
	}
	return out, nil
}

func (h ImageHandler) progress(w http.ResponseWriter, r *http.Request) {
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

func (h ImageHandler) registryLookup(c fuego.ContextNoBody) ([]registry.MetadataItem, error) {
	r := c.Request()
	name := r.URL.Query().Get("name")

	if name == "" {
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			id, err := uuid.Parse(idStr)
			if err != nil {
				return nil, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid image id"}
			}
			image, err := h.Images.FindByID(id)
			if err != nil {
				return nil, fuego.HTTPError{Status: http.StatusNotFound, Detail: "Image not found"}
			}
			name = image.RegistryName()
		}
	}

	if name == "" {
		return nil, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "name or id query parameter is required"}
	}

	entry := h.Resolver.ResolveEntry(name)
	if entry == nil {
		return []registry.MetadataItem{}, nil
	}
	meta := make([]registry.MetadataItem, len(entry.Metadata))
	copy(meta, entry.Metadata)
	return meta, nil
}

func imageToResponse(img *models.Image) dto.ImageResponse {
	var owner *dto.UserSummary
	if img.Owner != nil {
		owner = &dto.UserSummary{ID: img.Owner.ID, Email: img.Owner.Email}
	}
	return dto.ImageResponse{
		ID: img.ID, Name: img.Name, Tag: img.Tag,
		Title: img.Title, Description: img.Description,
		ThumbnailURL: img.ThumbnailURL, IsPublic: img.IsPublic,
		Status: img.Status, Error: img.Error,
		Metadata: img.Metadata, RegistryRef: img.RegistryRef,
		Owner:     owner,
		CreatedAt: img.CreatedAt, UpdatedAt: img.UpdatedAt,
	}
}

func mapImageError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fuego.HTTPError{Status: http.StatusNotFound, Detail: "Image not found"}
	}
	return fuego.HTTPError{Status: http.StatusInternalServerError, Detail: err.Error()}
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
