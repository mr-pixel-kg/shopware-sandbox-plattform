package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ImageHandler struct {
	images   *services.ImageService
	audit    *services.AuditService
	resolver RegistryResolver
}

type RegistryResolver interface {
	ResolveMetadata(imageName string) []registry.MetadataItem
}

func NewImageHandler(images *services.ImageService, audit *services.AuditService, resolver RegistryResolver) *ImageHandler {
	return &ImageHandler{images: images, audit: audit, resolver: resolver}
}

// ListPublic godoc
// @Summary      List public images
// @Description  Returns all images marked as public
// @Tags         Images
// @Produce      json
// @Success      200 {array} dto.ImageResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/public/images [get]
func (h *ImageHandler) ListPublic(c echo.Context) error {
	images, err := h.images.ListPublic()
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("IMAGE_LIST_FAILED", "Could not load public images").WithCause(err))
	}
	h.enrichMetadata(images)
	slog.Debug("listed public images", logging.RequestFields(c, "component", "image", "count", len(images))...)
	return c.JSON(200, toImageResponses(images))
}

// ListAll godoc
// @Summary      List all images
// @Description  Returns all images including private ones
// @Tags         Images
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} dto.ImageResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/images [get]
func (h *ImageHandler) ListAll(c echo.Context) error {
	images, err := h.images.ListAll()
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("IMAGE_LIST_FAILED", "Could not load images").WithCause(err))
	}
	h.enrichMetadata(images)
	slog.Debug("listed all images", logging.RequestFields(c, "component", "image", "count", len(images))...)
	return c.JSON(200, toImageResponses(images))
}

// Create godoc
// @Summary      Create an image
// @Description  Register a new Docker image. If not available locally, a background pull is started.
// @Tags         Images
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateImageRequest true "Image details"
// @Success      201 {object} models.Image
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Router       /api/images [post]
func (h *ImageHandler) Create(c echo.Context) error {
	var input dto.CreateImageRequest
	if err := bindAndValidate(c, &input); err != nil {
		return responses.FromError(c, err)
	}

	auth := mw.MustAuth(c)
	slog.Debug("image creation requested", logging.RequestFields(c,
		"component", "image",
		"user_id", auth.UserID.String(),
		"name", input.Name,
		"tag", input.Tag,
		"is_public", input.IsPublic,
	)...)

	metadataJSON, _ := json.Marshal(input.Metadata)
	image, err := h.images.CreateForUser(
		c.Request().Context(),
		&auth.UserID,
		input.Name,
		input.Tag,
		input.Title,
		input.Description,
		input.IsPublic,
		metadataJSON,
		nil,
	)
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("IMAGE_CREATE_FAILED", err.Error()).WithCause(err))
	}

	_ = h.audit.Log(&auth.UserID, "image.created", c.RealIP(), map[string]any{
		"name": input.Name,
		"tag":  input.Tag,
	})

	slog.Info("image created", logging.RequestFields(c,
		"component", "image",
		"user_id", auth.UserID.String(),
		"image_id", image.ID.String(),
		"image", image.FullName(),
		"status", image.Status,
	)...)
	h.enrichMetadata([]models.Image{*image})
	return c.JSON(201, toImageResponse(image))
}

// Update godoc
// @Summary      Update an image
// @Description  Update image metadata and visibility
// @Tags         Images
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Image ID" format(uuid)
// @Param        body body dto.UpdateImageRequest true "Updated image fields"
// @Success      200 {object} models.Image
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/images/{id} [put]
func (h *ImageHandler) Update(c echo.Context) error {
	auth := mw.MustAuth(c)
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid image id")
	if err != nil {
		return responses.FromError(c, err)
	}

	var input dto.UpdateImageRequest
	if err := bindAndValidate(c, &input); err != nil {
		return responses.FromError(c, err)
	}

	slog.Debug("image update requested", logging.RequestFields(c,
		"component", "image",
		"user_id", auth.UserID.String(),
		"image_id", id.String(),
		"is_public", input.IsPublic,
	)...)
	metadataJSON, _ := json.Marshal(input.Metadata)
	image, err := h.images.Update(id, input.Title, input.Description, input.IsPublic, metadataJSON)
	if err != nil {
		return mapImageError(c, "IMAGE_UPDATE_FAILED", "Could not update image", err)
	}

	slog.Info("image updated", logging.RequestFields(c,
		"component", "image",
		"user_id", auth.UserID.String(),
		"image_id", image.ID.String(),
		"is_public", image.IsPublic,
		"has_thumbnail", image.ThumbnailURL != nil,
	)...)
	_ = h.audit.Log(&auth.UserID, "image.updated", c.RealIP(), map[string]any{"imageId": image.ID.String()})
	h.enrichMetadata([]models.Image{*image})
	return c.JSON(http.StatusOK, toImageResponse(image))
}

// UploadThumbnail godoc
// @Summary      Upload an image thumbnail
// @Description  Upload or replace the thumbnail for an image
// @Tags         Images
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "Image ID" format(uuid)
// @Param        thumbnail formData file true "Thumbnail file"
// @Success      200 {object} models.Image
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/images/{id}/thumbnail [post]
func (h *ImageHandler) UploadThumbnail(c echo.Context) error {
	auth := mw.MustAuth(c)
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid image id")
	if err != nil {
		return responses.FromError(c, err)
	}

	fileHeader, err := c.FormFile("thumbnail")
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Missing thumbnail upload"))
	}

	file, err := fileHeader.Open()
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("THUMBNAIL_UPLOAD_FAILED", "Could not open thumbnail upload").WithCause(err))
	}
	defer file.Close()

	slog.Debug("thumbnail upload requested", logging.RequestFields(c,
		"component", "image",
		"user_id", auth.UserID.String(),
		"image_id", id.String(),
		"filename", fileHeader.Filename,
		"size", fileHeader.Size,
	)...)
	image, err := h.images.SaveThumbnail(id, file, fileHeader.Filename, fileHeader.Header.Get(echo.HeaderContentType))
	if err != nil {
		if errors.Is(err, services.ErrUnsupportedThumbnailFormat) {
			return responses.FromAppError(c, apperror.BadRequest("THUMBNAIL_FORMAT_UNSUPPORTED", "Unsupported thumbnail format").WithCause(err))
		}
		return mapImageError(c, "THUMBNAIL_UPLOAD_FAILED", "Could not store thumbnail", err)
	}

	slog.Info("thumbnail uploaded", logging.RequestFields(c,
		"component", "image",
		"user_id", auth.UserID.String(),
		"image_id", image.ID.String(),
		"thumbnail_url", image.ThumbnailURL,
	)...)
	_ = h.audit.Log(&auth.UserID, "image.thumbnail_uploaded", c.RealIP(), map[string]any{"imageId": image.ID.String()})
	h.enrichMetadata([]models.Image{*image})
	return c.JSON(http.StatusOK, toImageResponse(image))
}

// DeleteThumbnail godoc
// @Summary      Delete an image thumbnail
// @Description  Remove the thumbnail associated with an image
// @Tags         Images
// @Security     BearerAuth
// @Param        id path string true "Image ID" format(uuid)
// @Success      204
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/images/{id}/thumbnail [delete]
func (h *ImageHandler) DeleteThumbnail(c echo.Context) error {
	auth := mw.MustAuth(c)
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid image id")
	if err != nil {
		return responses.FromError(c, err)
	}

	slog.Debug("thumbnail deletion requested", logging.RequestFields(c, "component", "image", "user_id", auth.UserID.String(), "image_id", id.String())...)
	image, err := h.images.DeleteThumbnail(id)
	if err != nil {
		return mapImageError(c, "THUMBNAIL_DELETE_FAILED", "Could not delete thumbnail", err)
	}

	slog.Info("thumbnail deleted", logging.RequestFields(c,
		"component", "image",
		"user_id", auth.UserID.String(),
		"image_id", image.ID.String(),
	)...)
	_ = h.audit.Log(&auth.UserID, "image.thumbnail_deleted", c.RealIP(), map[string]any{"imageId": image.ID.String()})
	return c.NoContent(http.StatusNoContent)
}

// Delete godoc
// @Summary      Delete an image
// @Description  Remove a Docker image registration
// @Tags         Images
// @Security     BearerAuth
// @Param        id path string true "Image ID" format(uuid)
// @Success      204
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/images/{id} [delete]
func (h *ImageHandler) Delete(c echo.Context) error {
	auth := mw.MustAuth(c)
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid image id")
	if err != nil {
		return responses.FromError(c, err)
	}

	slog.Debug("image deletion requested", logging.RequestFields(c, "component", "image", "user_id", auth.UserID.String(), "image_id", id.String())...)
	if err := h.images.Delete(c.Request().Context(), id); err != nil {
		return responses.FromAppError(c, apperror.Internal("IMAGE_DELETE_FAILED", "Could not delete image").WithCause(err))
	}

	slog.Info("image deleted", logging.RequestFields(c, "component", "image", "user_id", auth.UserID.String(), "image_id", id.String())...)
	_ = h.audit.Log(&auth.UserID, "image.deleted", c.RealIP(), map[string]any{"imageId": id.String()})
	return c.NoContent(204)
}

// ListPending godoc
// @Summary      List pending image operations
// @Description  Returns all images with ongoing operations with optional progress percentage
// @Tags         Images
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} dto.PendingImageResponse
// @Failure      401 {object} dto.ErrorResponse
// @Router       /api/images/pending [get]
func (h *ImageHandler) ListPending(c echo.Context) error {
	images, percents := h.images.ListPendingImages()

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
	return c.JSON(200, out)
}

// Progress godoc
// @Summary      Stream image progress
// @Description  SSE endpoint streaming progress events for image operations
// @Tags         Images
// @Produce      text/event-stream
// @Param        id path string true "Image ID" format(uuid)
// @Success      200 {object} dto.ImageProgressEvent "Last emitted SSE event payload"
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/images/{id}/progress [get]
func (h *ImageHandler) Progress(c echo.Context) error {
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid image id")
	if err != nil {
		return responses.FromError(c, err)
	}

	idStr := id.String()

	image, dbErr := h.images.FindByID(id)
	if dbErr != nil {
		return responses.FromAppError(c, apperror.NotFound("IMAGE_NOT_FOUND", "Image not found"))
	}

	switch image.Status {
	case models.ImageStatusReady:
		writeSSEHeaders(c)
		sendSSEEvent(c, map[string]any{"percent": 100, "status": "ready"})
		return nil

	case models.ImageStatusFailed:
		writeSSEHeaders(c)
		errMsg := ""
		if image.Error != nil {
			errMsg = *image.Error
		}
		sendSSEEvent(c, map[string]any{"percent": 0, "status": "failed", "error": errMsg})
		return nil

	case models.ImageStatusPulling:
		if !h.images.IsPulling(idStr) {
			writeSSEHeaders(c)
			sendSSEEvent(c, map[string]any{"percent": 0, "status": "failed", "error": "pull process not running"})
			return nil
		}

		writeSSEHeaders(c)
		ch, cancel := h.images.WatchPullProgress(idStr)
		defer cancel()

		ctx := c.Request().Context()
		for {
			select {
			case <-ctx.Done():
				return nil
			case progress, ok := <-ch:
				if !ok {
					return nil
				}
				sendSSEEvent(c, progress)
			}
		}

	case models.ImageStatusCommitting:
		writeSSEHeaders(c)
		sendSSEEvent(c, map[string]any{"percent": 0, "status": "committing"})
		return nil

	default:
		return responses.FromAppError(c, apperror.NotFound("IMAGE_NOT_FOUND", "Image not found"))
	}
}

// RegistryLookup godoc
// @Summary      Lookup registry metadata
// @Description  Return registry-defined metadata for an image by name or ID
// @Tags         Images
// @Produce      json
// @Param        name query string false "Image name (e.g. dockware/dev:6.6.9.0)"
// @Param        id query string false "Image ID" format(uuid)
// @Success      200 {array} registry.MetadataItem
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/registry/lookup [get]
func (h *ImageHandler) RegistryLookup(c echo.Context) error {
	name := c.QueryParam("name")

	if name == "" {
		if idStr := c.QueryParam("id"); idStr != "" {
			id, err := uuid.Parse(idStr)
			if err != nil {
				return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid image id"))
			}
			image, err := h.images.FindByID(id)
			if err != nil {
				return responses.FromAppError(c, apperror.NotFound("IMAGE_NOT_FOUND", "Image not found"))
			}
			name = image.RegistryName()
		}
	}

	if name == "" {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "name or id query parameter is required"))
	}

	meta := h.resolver.ResolveMetadata(name)
	return c.JSON(http.StatusOK, meta)
}

func (h *ImageHandler) enrichMetadata(images []models.Image) {
	for i := range images {
		reg := h.resolver.ResolveMetadata(images[i].RegistryName())
		merged := mergeRegistryAndDB(reg, images[i].Metadata)
		data, _ := json.Marshal(merged)
		images[i].Metadata = datatypes.JSON(data)
	}
}

func mapImageError(c echo.Context, code, message string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.FromAppError(c, apperror.NotFound("IMAGE_NOT_FOUND", "Image not found").WithCause(err))
	}

	return responses.FromAppError(c, apperror.Internal(code, message).WithCause(err))
}
