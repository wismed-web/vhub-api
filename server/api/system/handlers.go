package system

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// *** after implementing, register with path in 'system.go' *** //

// @Title api service version
// @Summary get this api service version
// @Description
// @Tags    System
// @Accept  json
// @Produce json
// @Success 200 "OK - get its version"
// @Router /api/system/ver [get]
func Ver(c echo.Context) error {
	return c.JSON(http.StatusOK, version)
}

// @Title api service tag
// @Summary get this api service project github version tag
// @Description
// @Tags    System
// @Accept  json
// @Produce json
// @Success 200 "OK - get its tag"
// @Router /api/system/tag [get]
func Tag(c echo.Context) error {
	return c.JSON(http.StatusOK, tag)
}