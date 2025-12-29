package interfaces

import "github.com/labstack/echo/v4"

// Controller defines the interface that all controllers should implement
type Controller interface {
	RegisterRoutes(e *echo.Group)
}


