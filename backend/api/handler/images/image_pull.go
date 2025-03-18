package images

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type PullImageRequest struct {
	ImageName string `json:"image_name" example:"dockware/dev"`
	ImageTag  string `json:"image_tag" example:"6.6.8.2"`
}

type PullImageResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Docker Image created successfully"`
}

// PullImageHandler pulls/registers a docker image into the system
// @Summary Pull Docker Image
// @Description Pulls a docker image and register it into the system
// @Tags Docker Image Management
// @Accept json
// @Produce json
// @Param image body PullImageRequest true "Image Input"
// @Success 200 {object} PullImageResponse
// @Failure 400 {object} map[string]string
// @Security BasicAuth
// @Router /api/images [post]
func (h *ImageHandler) PullImageHandler(c echo.Context) error {
	ctx := c.Request().Context()
	var input PullImageRequest

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request format",
		})
	}

	if len(input.ImageName) < 5 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Image name must be at least 5 characters long",
		})
	}

	imageName := input.ImageName + ":" + input.ImageTag
	h.ImageService.PullImage(ctx, imageName)

	output := PullImageResponse{
		Message: "Image " + input.ImageName + ":" + input.ImageTag + " created successfully",
		Status:  "success",
	}

	return c.JSON(http.StatusOK, output)
}
