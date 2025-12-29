package officers

import (
	"police-trafic-api-frontend-aligned/internal/shared/responses"

	"github.com/labstack/echo/v4"
)

// Controller handles officers HTTP requests for mobile app
type Controller struct {
	service Service
}

// NewController creates a new officers controller
func NewController(service Service) *Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes registers officers routes on the group
func (ctrl *Controller) RegisterRoutes(e *echo.Group) {
	officers := e.Group("/officers")

	officers.GET("/:id/dashboard", ctrl.GetOfficerDashboard)
	officers.GET("/:id/statistics", ctrl.GetOfficerStatistics)
}

// GetOfficerDashboard handles GET /officers/:id/dashboard
func (ctrl *Controller) GetOfficerDashboard(c echo.Context) error {
	officerID := c.Param("id")
	if officerID == "" {
		return responses.BadRequest(c, "Officer ID is required")
	}

	period := c.QueryParam("period")
	if period == "" {
		period = "daily"
	}

	dashboard, err := ctrl.service.GetOfficerDashboard(c.Request().Context(), officerID, period)
	if err != nil {
		return responses.NotFound(c, "Officer not found")
	}

	return responses.Success(c, dashboard)
}

// GetOfficerStatistics handles GET /officers/:id/statistics
func (ctrl *Controller) GetOfficerStatistics(c echo.Context) error {
	officerID := c.Param("id")
	if officerID == "" {
		return responses.BadRequest(c, "Officer ID is required")
	}

	stats, err := ctrl.service.GetOfficerStatistics(c.Request().Context(), officerID)
	if err != nil {
		return responses.NotFound(c, "Officer not found")
	}

	return responses.Success(c, stats)
}
