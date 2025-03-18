package images

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

type ImageDeleteResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Docker Image removed successfully"`
}

// ImageDeleteHandler removes a docker image
// @Summary Remove Docker Image
// @Description Removes a docker image form the system
// @Tags Docker Image Management
// @Accept json
// @Produce json
// @Param id path string true "Image ID"
// @Success 200 {object} ImageDeleteResponse
// @Failure 400 {object} map[string]string
// @Security BasicAuth
// @Router /api/images/{id} [delete]
func (h *ImageHandler) ImageDeleteHandler(c echo.Context) error {

	ctx := c.Request().Context()
	imageId := c.Param("id")

	err := h.ImageService.DeleteImage(ctx, imageId)
	if err != nil {
		log.Printf("Failed to delete image: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	output := ImageDeleteResponse{
		Message: "Image " + imageId + " removed successfully",
		Status:  "success",
	}

	return c.JSON(http.StatusOK, output)
}
