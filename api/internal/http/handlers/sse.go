package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/labstack/echo/v4"
)

func writeSSEHeaders(c echo.Context) {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(200)
}

func sendSSEEvent(c echo.Context, v any) {
	data, _ := json.Marshal(v)
	fmt.Fprintf(c.Response(), "data: %s\n\n", data)
	c.Response().Flush()
}
