package handlers

import (
	"encoding/json"
	"fmt"
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
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"gorm.io/gorm"
)

type ImageHandler struct {
	images *services.ImageService
	audit  *services.AuditService
}

func NewImageHandler(images *services.ImageService, audit *services.AuditService) *ImageHandler {
	return &ImageHandler{images: images, audit: audit}
}

// ListPublic godoc
// @Summary      List public images
// @Description  Returns all images marked as public
// @Tags         Images
// @Produce      json
// @Success      200 {array} models.Image
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/public/images [get]
func (h *ImageHandler) ListPublic(c echo.Context) error {
	images, err := h.images.ListPublic()
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("IMAGE_LIST_FAILED", "Could not load public images").WithCause(err))
	}
	slog.Info("listed public images", logging.RequestFields(c, "count", len(images))...)
	return c.JSON(200, images)
}

// ListAll godoc
// @Summary      List all images
// @Description  Returns all images including private ones
// @Tags         Images
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} models.Image
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/images [get]
func (h *ImageHandler) ListAll(c echo.Context) error {
	images, err := h.images.ListAll()
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("IMAGE_LIST_FAILED", "Could not load images").WithCause(err))
	}
	slog.Info("listed all images", logging.RequestFields(c, "count", len(images))...)
	return c.JSON(200, images)
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
// @Success      202 {object} dto.PendingPullResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Router       /api/images [post]
func (h *ImageHandler) Create(c echo.Context) error {
	var input dto.CreateImageRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	auth := mw.MustAuth(c)
	slog.Info("image creation requested", logging.RequestFields(c,
		"user_id", auth.UserID.String(),
		"name", input.Name,
		"tag", input.Tag,
		"is_public", input.IsPublic,
	)...)

	image, pending, err := h.images.CreateForUser(
		c.Request().Context(),
		&auth.UserID,
		input.Name,
		input.Tag,
		input.Title,
		input.Description,
		input.IsPublic,
	)
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("IMAGE_CREATE_FAILED", err.Error()).WithCause(err))
	}

	_ = h.audit.Log(&auth.UserID, "image.created", c.RealIP(), map[string]any{
		"name": input.Name,
		"tag":  input.Tag,
	})

	// todo: think about it
	if pending != nil {
		slog.Info("image pull started", logging.RequestFields(c,
			"user_id", auth.UserID.String(),
			"image_id", pending.ID.String(),
			"image", input.Name+":"+input.Tag,
		)...)
		return c.JSON(202, dto.PendingPullResponse{
			ID:      pending.ID.String(),
			Name:    pending.Name,
			Tag:     pending.Tag,
			Title:   pending.Title,
			Percent: 0,
			Status:  "pulling",
		})
	}

	slog.Info("image created", logging.RequestFields(c,
		"user_id", auth.UserID.String(),
		"image_id", image.ID.String(),
		"image", image.FullName(),
	)...)
	return c.JSON(201, image)
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
func (h *ImageHandler) Update(c echo.Context) error {
	auth := mw.MustAuth(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid image id"))
	}

	var input dto.UpdateImageRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	slog.Info("image update requested", logging.RequestFields(c,
		"user_id", auth.UserID.String(),
		"image_id", id.String(),
		"is_public", input.IsPublic,
	)...)
	image, err := h.images.Update(id, input.Title, input.Description, input.IsPublic)
	if err != nil {
		return mapImageError(c, "IMAGE_UPDATE_FAILED", "Could not update image", err)
	}

	slog.Info("image updated successfully", logging.RequestFields(c,
		"user_id", auth.UserID.String(),
		"image_id", image.ID.String(),
		"is_public", image.IsPublic,
		"has_thumbnail", image.ThumbnailURL != nil,
	)...)
	_ = h.audit.Log(&auth.UserID, "image.updated", c.RealIP(), map[string]any{"imageId": image.ID.String()})
	return c.JSON(http.StatusOK, image)
}

func (h *ImageHandler) UploadThumbnail(c echo.Context) error {
	auth := mw.MustAuth(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid image id"))
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

	slog.Info("thumbnail upload requested", logging.RequestFields(c,
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

	slog.Info("thumbnail uploaded successfully", logging.RequestFields(c,
		"user_id", auth.UserID.String(),
		"image_id", image.ID.String(),
		"thumbnail_url", image.ThumbnailURL,
	)...)
	_ = h.audit.Log(&auth.UserID, "image.thumbnail_uploaded", c.RealIP(), map[string]any{"imageId": image.ID.String()})
	return c.JSON(http.StatusOK, image)
}

func (h *ImageHandler) DeleteThumbnail(c echo.Context) error {
	auth := mw.MustAuth(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid image id"))
	}

	slog.Info("thumbnail deletion requested", logging.RequestFields(c, "user_id", auth.UserID.String(), "image_id", id.String())...)
	image, err := h.images.DeleteThumbnail(id)
	if err != nil {
		return mapImageError(c, "THUMBNAIL_DELETE_FAILED", "Could not delete thumbnail", err)
	}

	slog.Info("thumbnail deleted successfully", logging.RequestFields(c,
		"user_id", auth.UserID.String(),
		"image_id", image.ID.String(),
		"has_thumbnail", image.ThumbnailURL != nil,
	)...)
	_ = h.audit.Log(&auth.UserID, "image.thumbnail_deleted", c.RealIP(), map[string]any{"imageId": image.ID.String()})
	return c.NoContent(http.StatusNoContent)
}

func (h *ImageHandler) Delete(c echo.Context) error {
	auth := mw.MustAuth(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid image id"))
	}

	slog.Info("image deletion requested", logging.RequestFields(c, "user_id", auth.UserID.String(), "image_id", id.String())...)
	if err := h.images.Delete(c.Request().Context(), id); err != nil {
		return responses.FromAppError(c, apperror.Internal("IMAGE_DELETE_FAILED", "Could not delete image").WithCause(err))
	}

	slog.Info("image deleted successfully", logging.RequestFields(c, "user_id", auth.UserID.String(), "image_id", id.String())...)
	_ = h.audit.Log(&auth.UserID, "image.deleted", c.RealIP(), map[string]any{"imageId": id.String()})
	return c.NoContent(204)
}

// ListPulls godoc
// @Summary      List ongoing image pulls
// @Description  Returns all images currently being pulled (in-memory only, not persisted)
// @Tags         Images
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} dto.PendingPullResponse
// @Failure      401 {object} dto.ErrorResponse
// @Router       /api/images/pulls [get]
func (h *ImageHandler) ListPulls(c echo.Context) error {
	pending := h.images.ListPendingPulls()

	out := make([]dto.PendingPullResponse, len(pending))
	for i, p := range pending {
		out[i] = dto.PendingPullResponse{
			ID:      p.ID.String(),
			Name:    p.Name,
			Tag:     p.Tag,
			Title:   p.Title,
			Percent: p.Percent,
			Status:  p.Status,
		}
	}
	return c.JSON(200, out)
}

// PullProgress godoc
// @Summary      Stream image pull progress
// @Description  SSE endpoint streaming pull progress events. Each event is JSON with "percent" (int) and "status" (string) fields.
// @Tags         Images
// @Produce      text/event-stream
// @Param        id path string true "Image ID" format(uuid)
// @Success      200 {object} dto.ImagePullProgressEvent "Last emitted SSE event payload"
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/images/{id}/progress [get]
func (h *ImageHandler) PullProgress(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid image id"))
	}

	idStr := id.String()

	// If the image already exists in the db, its done and return ready
	if _, dbErr := h.images.FindByID(id); dbErr == nil {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().WriteHeader(200)
		data, _ := json.Marshal(map[string]any{"percent": 100, "status": "ready"})
		fmt.Fprintf(c.Response(), "data: %s\n\n", data)
		c.Response().Flush()
		return nil
	}

	// if its not a pending pull too, return 404.
	if !h.images.IsPulling(idStr) {
		return responses.FromAppError(c, apperror.NotFound("IMAGE_NOT_FOUND", "Image not found"))
	}

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(200)

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
			data, _ := json.Marshal(progress)
			fmt.Fprintf(c.Response(), "data: %s\n\n", data)
			c.Response().Flush()
		}
	}
}

func mapImageError(c echo.Context, code, message string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return responses.FromAppError(c, apperror.NotFound("IMAGE_NOT_FOUND", "Image not found").WithCause(err))
	}

	return responses.FromAppError(c, apperror.Internal(code, message).WithCause(err))
}
