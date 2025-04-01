package sandboxes

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

type SandboxSnapshotCreateRequest struct {
	ImageName string `json:"image_name" example:"mr-pixel/shopware-demoshop:6.6.8.2"`
}

// SandboxSnapshotHandler creates a new sandbox image from the current sandbox container
// @Summary Creates a new sandbox image from the sandbox container
// @Description Creates a new sandbox image from the sandbox container
// @Tags Sandbox Management
// @Accept json
// @Produce json
// @Param id path string true "Sandbox ID" example(67777b4e-946f-4462-b689-3c608d2d7938)
// @Param image body SandboxSnapshotCreateRequest true "Image Input"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security BasicAuth
// @Router /api/sandboxes/{id}/snapshot [post]
func (h *SandboxHandler) SandboxSnapshotHandler(c echo.Context) error {

	ctx := c.Request().Context()
	sandboxId := c.Param("id")

	var input SandboxCreateRequest

	// Parse input
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request format",
		})
	}

	snapshotImageName := input.ImageName
	if snapshotImageName == "" {
		snapshotImageName = "mr-pixel/sandbox-snapshot:" + uuid.New().String()
	}

	image, err := h.SandboxService.CommitSandbox(ctx, sandboxId, snapshotImageName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message":   "Sandbox snapshot created successfully",
		"sandboxId": sandboxId,
		"image":     image.ImageName + ":" + image.ImageTag,
		"imageId":   image.ID,
	})
}
