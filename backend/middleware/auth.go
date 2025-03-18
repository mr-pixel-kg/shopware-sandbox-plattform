package middleware // AuthMiddleware prüft Benutzer anhand der DB

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/config"
	"log"
)

func AuthMiddleware(authConfig config.AuthConfig) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {

		// Check password
		if username != authConfig.Username || password != authConfig.Password {
			log.Printf("Wrong login credentials for user %s\n", username)
			return false, nil // Wrong credentials
		}

		// Store user info
		c.Set("username", username)
		return true, nil
	})
}
