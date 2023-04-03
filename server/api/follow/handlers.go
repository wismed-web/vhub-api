package follow

import (
	"net/http"

	em "github.com/digisan/event-mgr"
	"github.com/labstack/echo/v4"
)

// @Title get a Post follower-Post id
// @Summary get a specified Post follower-Post id group.
// @Description
// @Tags    Follow
// @Accept  json
// @Produce json
// @Param   followee query string true "followee Post ID"
// @Success 200 "OK - get successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/follow/follower [get]
// @Security ApiKeyAuth
func Followers(c echo.Context) error {
	var (
		flwee = c.QueryParam("followee")
	)

	flwers, err := em.Followers(flwee)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	// if len(flwers) == 0 {
	// 	return c.JSON(http.StatusNotFound, flwers)
	// }
	return c.JSON(http.StatusOK, flwers)
}
