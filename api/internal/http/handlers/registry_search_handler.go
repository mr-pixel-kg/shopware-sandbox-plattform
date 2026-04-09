package handlers

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
)

type RegistrySearchHandler struct {
	Search *services.RegistrySearchService
}

func (h RegistrySearchHandler) SearchImages(c fuego.ContextNoBody) (dto.RegistryImageSearchResponse, error) {
	q := c.Request().URL.Query().Get("q")
	if len(q) < 2 {
		return dto.RegistryImageSearchResponse{}, fuego.HTTPError{
			Status: http.StatusBadRequest,
			Detail: "Query must be at least 2 characters",
		}
	}

	results, err := h.Search.SearchImages(c.Request().Context(), q)
	if err != nil {
		return dto.RegistryImageSearchResponse{}, fuego.HTTPError{
			Status: http.StatusInternalServerError,
			Detail: "Image search failed",
		}
	}

	return dto.RegistryImageSearchResponse{Results: results}, nil
}

func (h RegistrySearchHandler) SearchTags(c fuego.ContextNoBody) (dto.RegistryTagSearchResponse, error) {
	image := c.Request().URL.Query().Get("image")
	if image == "" {
		return dto.RegistryTagSearchResponse{}, fuego.HTTPError{
			Status: http.StatusBadRequest,
			Detail: "image query parameter is required",
		}
	}

	q := c.Request().URL.Query().Get("q")

	results, err := h.Search.SearchTags(c.Request().Context(), image, q)
	if err != nil {
		return dto.RegistryTagSearchResponse{}, fuego.HTTPError{
			Status: http.StatusInternalServerError,
			Detail: "Tag search failed",
		}
	}

	return dto.RegistryTagSearchResponse{Results: results}, nil
}
