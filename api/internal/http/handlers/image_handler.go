package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type ImageHandler struct {
	images *services.ImageService
	audit  *services.AuditService
}

func NewImageHandler(images *services.ImageService, audit *services.AuditService) *ImageHandler {
	return &ImageHandler{images: images, audit: audit}
}

func (h *ImageHandler) ListPublic(c echo.Context) error {
	images, err := h.images.ListPublic()
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "IMAGE_LIST_FAILED", "Could not load public images")
	}
	return c.JSON(http.StatusOK, images)
}

func (h *ImageHandler) ListAll(c echo.Context) error {
	images, err := h.images.ListAll()
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "IMAGE_LIST_FAILED", "Could not load images")
	}
	return c.JSON(http.StatusOK, images)
}

func (h *ImageHandler) Create(c echo.Context) error {
	var input dto.CreateImageRequest
	if err := c.Bind(&input); err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	auth := mw.MustAuth(c)
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
		return responses.Error(c, http.StatusBadRequest, "IMAGE_CREATE_FAILED", err.Error())
	}

	_ = h.audit.Log(&auth.UserID, "image.created", c.RealIP(), map[string]any{"imageId": image.ID.String()})
	return c.JSON(http.StatusCreated, image)
}

func (h *ImageHandler) Delete(c echo.Context) error {
	auth := mw.MustAuth(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid image id")
	}

	if err := h.images.Delete(c.Request().Context(), id); err != nil {
		return responses.Error(c, http.StatusInternalServerError, "IMAGE_DELETE_FAILED", "Could not delete image")
	}

	_ = h.audit.Log(&auth.UserID, "image.deleted", c.RealIP(), map[string]any{"imageId": id.String()})
	return c.NoContent(http.StatusNoContent)
}
