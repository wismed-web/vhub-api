package bookmark

import (
	"net/http"

	em "github.com/digisan/event-mgr"
	u "github.com/digisan/user-mgr/user"
	"github.com/labstack/echo/v4"
)

// @Title toggle a bookmark for a post
// @Summary add or remove a personal bookmark for a post.
// @Description
// @Tags    Bookmark
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for toggling a bookmark"
// @Success 200 "OK - toggled bookmark successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/bookmark/toggle/{id} [patch]
// @Security ApiKeyAuth
func Bookmark(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname = invoker.UName
		id    = c.Param("id")
	)
	bm, err := em.NewBookmark(uname, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	has, err := bm.ToggleEvent(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, has)
}

// @Title get current user's bookmark status for a post
// @Summary get current login user's bookmark status for a post.
// @Description
// @Tags    Bookmark
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for checking bookmark status"
// @Success 200 "OK - get bookmark status successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/bookmark/status/{id} [get]
// @Security ApiKeyAuth
func Status(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname = invoker.UName
		id    = c.Param("id")
	)
	bm, err := em.NewBookmark(uname, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, bm.HasEvent(id))
}

// @Title get bookmarked Posts
// @Summary get all bookmarked Post ids.
// @Description
// @Tags    Bookmark
// @Accept  json
// @Produce json
// @Param   order query string false "order[desc asc] to get Post ids ordered by event time"
// @Success 200 "OK - get successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/bookmark/marked [get]
// @Security ApiKeyAuth
func Marked(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname = invoker.UName
		order = c.QueryParam("order")
	)

	bm, err := em.NewBookmark(uname, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, bm.Bookmarks(order))
}
