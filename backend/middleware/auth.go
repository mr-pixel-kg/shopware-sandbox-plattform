package middleware // AuthMiddleware prüft Benutzer anhand der DB

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/config"
	"log"
)

func AuthRequiredMiddleware(authConfig config.AuthConfig) echo.MiddlewareFunc {
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

func OptionalAuthMiddleware(authConfig config.AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				// No auth header, allow to continue
				//log.Println("Request: No auth header found")
				return next(c)
			}

			// Parse basic auth header
			username, password, ok := c.Request().BasicAuth()
			if !ok || username != authConfig.Username || password != authConfig.Password {
				// Wrong credentials, but allow to continue
				//log.Printf("Request: User %s has wrong credentials\n", username)
				return next(c)
			}

			// User is validated
			//log.Printf("Request: User %s is validated\n", username)
			c.Set("username", username)
			return next(c)
		}
	}
}

func IsUserLoggedIn(context echo.Context) bool {
	if context.Get("username") != nil {
		//log.Println("User is logged in")
		return true
	}
	//log.Println("User is not logged in")
	return false
}

func GetCurrentUserName(context echo.Context) *string {
	username, ok := context.Get("username").(string)
	if !ok {
		return nil
	}
	return &username
}
