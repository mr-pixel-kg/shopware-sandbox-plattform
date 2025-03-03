package sandboxes

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

type SandboxCreateRequest struct {
	ImageName string `json:"image_name" example:"dockware/dev:6.6.8.2"`
	Lifetime  int    `json:"lifetime" example:"1440"`
}

type SandboxCreateResponse struct {
	Status        string `json:"status" example:"success"`
	Message       string `json:"message" example:"Sandbox created successfully"`
	Image         string `json:"image" example:"dockware/dev:6.6.8.2"`
	ContainerName string `json:"container_name" example:"sandbox-67777b4e-946f-4462-b689-3c608d2d7938"`
	ContainerId   string `json:"container_id" example:"9a7f95b73018432cb88ebed68046c59a4bed05b2abc809f6fbf39a1173c06ac9"`
	Url           string `json:"url" example:"https://sandbox-67777b4e-946f-4462-b689-3c608d2d7938.shopshredder.zion.mr-pixel.de"`
	SandboxId     string `json:"sandbox_id" example:"67777b4e-946f-4462-b689-3c608d2d7938"`
}

// SandboxCreateHandler creates a new sandbox for requested image
// @Summary Create new sandbox
// @Description Creates a new sandbox docker container
// @Tags Sandbox Management
// @Accept json
// @Produce json
// @Param image body SandboxCreateRequest true "Image Input"
// @Success 200 {object} SandboxCreateResponse
// @Failure 400 {object} map[string]string
// @Router /api/sandboxes [post]
func (h *SandboxHandler) SandboxCreateHandler(c echo.Context) error {

	ctx := c.Request().Context()
	var input SandboxCreateRequest

	// Parse input
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request format",
		})
	}

	// Input validation
	if input.Lifetime > 1440 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Lifetime should be less than 1440",
		})
	}

	if input.Lifetime < 5 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Lifetime should be greater than 5",
		})
	}

	imageName := input.ImageName

	sandbox, err := h.SandboxService.CreateSandbox(ctx, imageName, input.Lifetime)
	if err != nil {
		log.Printf("Failed to create sandbox %s: %v", imageName, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create sandbox environment")
	}

	output := SandboxCreateResponse{
		Message:       "Sandbox created successfully",
		Status:        "success",
		ContainerName: sandbox.ContainerName,
		ContainerId:   sandbox.ContainerID,
		Image:         imageName,
		Url:           sandbox.URL,
		SandboxId:     sandbox.ID,
	}

	return c.JSON(http.StatusOK, output)
}
