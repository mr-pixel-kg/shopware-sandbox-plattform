package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type SandboxHandler struct {
	sandboxes *services.SandboxService
}

func NewSandboxHandler(sandboxes *services.SandboxService) *SandboxHandler {
	return &SandboxHandler{sandboxes: sandboxes}
}

func (h *SandboxHandler) List(c echo.Context) error {
	sandboxes, err := h.sandboxes.ListActive()
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "SANDBOX_LIST_FAILED", "Could not load sandboxes")
	}
	return c.JSON(http.StatusOK, sandboxes)
}

func (h *SandboxHandler) ListMine(c echo.Context) error {
	auth := mw.MustAuth(c)
	sandboxes, err := h.sandboxes.ListByUser(auth.UserID)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "SANDBOX_LIST_FAILED", "Could not load own sandboxes")
	}
	return c.JSON(http.StatusOK, sandboxes)
}

func (h *SandboxHandler) ListGuest(c echo.Context) error {
	guest := mw.MustGuest(c)
	sandboxes, err := h.sandboxes.ListByGuestSession(guest.SessionID)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "SANDBOX_LIST_FAILED", "Could not load guest sandboxes")
	}
	return c.JSON(http.StatusOK, sandboxes)
}

func (h *SandboxHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid sandbox id")
	}

	sandbox, err := h.sandboxes.FindByID(id)
	if err != nil {
		return responses.Error(c, http.StatusNotFound, "SANDBOX_NOT_FOUND", "Sandbox not found")
	}
	return c.JSON(http.StatusOK, sandbox)
}

func (h *SandboxHandler) CreatePublicDemo(c echo.Context) error {
	var input dto.CreateSandboxRequest
	if err := c.Bind(&input); err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	imageID, err := uuid.Parse(input.ImageID)
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid image id")
	}

	guest := mw.MustGuest(c)
	sandbox, err := h.sandboxes.Create(c.Request().Context(), services.CreateSandboxInput{
		ImageID:        imageID,
		GuestSessionID: &guest.SessionID,
		ClientIP:       c.RealIP(),
	})
	if err != nil {
		return mapSandboxError(c, err)
	}

	return c.JSON(http.StatusCreated, sandbox)
}

func (h *SandboxHandler) CreatePrivateSandbox(c echo.Context) error {
	var input dto.CreateSandboxRequest
	if err := c.Bind(&input); err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	imageID, err := uuid.Parse(input.ImageID)
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid image id")
	}

	auth := mw.MustAuth(c)
	var ttl *time.Duration
	if input.TTLMinutes != nil {
		duration := time.Duration(*input.TTLMinutes) * time.Minute
		ttl = &duration
	}

	sandbox, err := h.sandboxes.Create(c.Request().Context(), services.CreateSandboxInput{
		ImageID:  imageID,
		UserID:   &auth.UserID,
		ClientIP: c.RealIP(),
		TTL:      ttl,
	})
	if err != nil {
		return mapSandboxError(c, err)
	}

	return c.JSON(http.StatusCreated, sandbox)
}

func (h *SandboxHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid sandbox id")
	}

	auth := mw.MustAuth(c)
	if err := h.sandboxes.Delete(c.Request().Context(), id, c.RealIP(), &auth.UserID); err != nil {
		return mapSandboxError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *SandboxHandler) Snapshot(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid sandbox id")
	}

	var input dto.CreateSnapshotRequest
	if err := c.Bind(&input); err != nil {
		return responses.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	auth := mw.MustAuth(c)
	image, err := h.sandboxes.CreateSnapshot(c.Request().Context(), services.CreateSnapshotInput{
		SandboxID:    id,
		Name:         input.Name,
		Tag:          input.Tag,
		Title:        input.Title,
		Description:  input.Description,
		ThumbnailURL: input.ThumbnailURL,
		IsPublic:     input.IsPublic,
		ClientIP:     c.RealIP(),
		UserID:       &auth.UserID,
	})
	if err != nil {
		return mapSandboxError(c, err)
	}

	return c.JSON(http.StatusCreated, image)
}

func mapSandboxError(c echo.Context, err error) error {
	switch err {
	case services.ErrSandboxLimitReached:
		return responses.Error(c, http.StatusConflict, "SANDBOX_LIMIT_REACHED", "Maximum number of sandboxes reached")
	case services.ErrSandboxNotFound:
		return responses.Error(c, http.StatusNotFound, "SANDBOX_NOT_FOUND", "Sandbox not found")
	default:
		return responses.Error(c, http.StatusInternalServerError, "SANDBOX_ERROR", err.Error())
	}
}
