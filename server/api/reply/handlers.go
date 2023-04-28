package reply

import (
	"net/http"

	em "github.com/digisan/event-mgr"
	. "github.com/digisan/go-generics/v2"
	"github.com/labstack/echo/v4"
)

// @Title   get Post Replies
// @Summary get specified Post Reply id group.
// @Description
// @Tags    Reply
// @Accept  json
// @Produce json
// @Param   pid path string true "Post ID"
// @Success 200 "OK - get successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/reply/{pid} [get]
// @Security ApiKeyAuth
func Replies(c echo.Context) error {
	var (
		pid = c.Param("pid")
	)
	replies, err := em.Followers(pid)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, Reverse(replies))
}
