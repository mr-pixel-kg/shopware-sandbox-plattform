package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
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
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
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
	image, err := h.images.CreateForUser(
		c.Request().Context(),
		&auth.UserID,
		input.Name,
		input.Tag,
		input.Title,
		input.Description,
		input.ThumbnailURL,
		input.IsPublic,
	)
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("IMAGE_CREATE_FAILED", err.Error()).WithCause(err))
	}

	slog.Info("image created successfully", logging.RequestFields(c,
		"user_id", auth.UserID.String(),
		"image_id", image.ID.String(),
		"image", image.FullName(),
		"is_public", image.IsPublic,
	)...)
	_ = h.audit.Log(&auth.UserID, "image.created", c.RealIP(), map[string]any{"imageId": image.ID.String()})
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

	img, err := h.images.FindByID(id)
	if err != nil {
		return responses.FromAppError(c, apperror.NotFound("IMAGE_NOT_FOUND", "Image not found"))
	}

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(200)

	if img.Status == "ready" || img.Status == "failed" {
		data, _ := json.Marshal(map[string]any{
			"percent": 100,
			"status":  img.Status,
		})
		fmt.Fprintf(c.Response(), "data: %s\n\n", data)
		c.Response().Flush()
		return nil
	}

	ch, cancel := h.images.WatchPullProgress(id.String())
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
