package manage

import (
	"fmt"
	"net/http"
	"time"

	em "github.com/digisan/event-mgr"
	. "github.com/digisan/go-generics/v2"
	lk "github.com/digisan/logkit"
	u "github.com/digisan/user-mgr/user"
	"github.com/labstack/echo/v4"
)

// @Title   delete Post
// @Summary delete one Post content.
// @Description
// @Tags    Post
// @Accept  json
// @Produce json
// @Param   id   query string true "Post ID for deleting"
// @Success 200 "OK - delete successfully"
// @Failure 400 "Fail - incorrect query param id"
// @Failure 404 "Fail - not found"
// @Failure 500 "Fail - internal error"
// @Router /api/post/delete/one [delete]
// @Security ApiKeyAuth
func DelOne(c echo.Context) error {
	var (
		id = c.QueryParam("id")
	)
	if len(id) == 0 {
		return c.String(http.StatusBadRequest, "'id' cannot be empty")
	}
	n, err := em.DelEvent(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, IF(n == 1, fmt.Sprintf("<%s> is deleted", id), fmt.Sprintf("<%s> is not existing, nothing to delete", id)))
}

// @Title erase one Post content
// @Summary erase one Post content permanently.
// @Description
// @Tags    Post
// @Accept  json
// @Produce json
// @Param   id   query string true "Post ID for erasing"
// @Success 200 "OK - erase successfully"
// @Failure 400 "Fail - incorrect query param id"
// @Failure 404 "Fail - not found"
// @Failure 500 "Fail - internal error"
// @Router /api/post/erase/one [delete]
// @Security ApiKeyAuth
func EraseOne(c echo.Context) error {
	var (
		id = c.QueryParam("id")
	)
	if len(id) == 0 {
		return c.String(http.StatusBadRequest, "'id' cannot be empty")
	}
	n, err := em.EraseEvents(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, IF(n == 1, fmt.Sprintf("<%s> is erased permanently", id), fmt.Sprintf("<%s> is not existing, nothing to erase", id)))
}

// @Title get own Post id group in a specific period
// @Summary get own Post id group in one specific time period.
// @Description
// @Tags    Post
// @Accept  json
// @Produce json
// @Param   period query string false "time period for query, format is 'yyyymm', e.g. '202206'. if missing, current yyyymm applies"
// @Success 200 "OK - get successfully"
// @Failure 400 "Fail - incorrect query param type"
// @Failure 404 "Fail - empty event ids"
// @Failure 500 "Fail - internal error"
// @Router /api/post/own/ids [get]
// @Security ApiKeyAuth
func OwnPosts(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname  = invoker.UName
		period = c.QueryParam("period")
	)

	if len(period) == 0 {
		period = time.Now().Format("200601")
	}
	if _, err := time.Parse("200601", period); err != nil {
		return c.String(http.StatusBadRequest, "'period' format must be 'yyyymm', e.g. '202206'")
	}

	lk.Log("%s -- %s", uname, period)

	ids, err := em.FetchOwn(uname, period)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	// if len(ids) == 0 {
	// 	return c.JSON(http.StatusNotFound, ids)
	// }
	return c.JSON(http.StatusOK, ids)
}

// @Title toggle a bookmark for a post
// @Summary add or remove a personal bookmark for a post.
// @Description
// @Tags    Post
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for toggling a bookmark"
// @Success 200 "OK - toggled bookmark successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/post/bookmark/{id} [patch]
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
// @Tags    Post
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for checking bookmark status"
// @Success 200 "OK - get bookmark status successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/post/bookmark/status/{id} [get]
// @Security ApiKeyAuth
func BookmarkStatus(c echo.Context) error {

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
// @Tags    Post
// @Accept  json
// @Produce json
// @Param   order query string false "order[desc asc] to get Post ids ordered by event time"
// @Success 200 "OK - get successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/post/bookmark/bookmarked [get]
// @Security ApiKeyAuth
func BookmarkedPosts(c echo.Context) error {

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

// @Title get a Post follower-Post ids
// @Summary get a specified Post follower-Post id group.
// @Description
// @Tags    Post
// @Accept  json
// @Produce json
// @Param   followee query string true "followee Post ID"
// @Success 200 "OK - get successfully"
// @Failure 404 "Fail - empty follower ids"
// @Failure 500 "Fail - internal error"
// @Router /api/post/follower/ids [get]
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

// @Title add or remove a thumbs-up for a post
// @Summary add or remove a personal thumbs-up for a post.
// @Description
// @Tags    Post
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for adding or removing thumbs-up"
// @Success 200 "OK - added or removed thumb successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/post/thumbs-up/{id} [patch]
// @Security ApiKeyAuth
func ThumbsUp(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname = invoker.UName
		id    = c.Param("id")
	)
	ep, err := em.NewEventParticipate(id, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	has, err := ep.TogglePtp("ThumbsUp", uname)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	ptps, err := ep.Ptps("ThumbsUp")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	lk.Log("---> %v", ptps)

	return c.JSON(http.StatusOK, struct {
		ThumbsUp bool
		Count    int
	}{
		has, len(ptps),
	})
}

// @Title get current user's thumbs-up status for a post
// @Summary get current login user's thumbs-up status for a post.
// @Description
// @Tags    Post
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for checking thumbs-up status"
// @Success 200 "OK - get thumbs-up status successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/post/thumbs-up/status/{id} [get]
// @Security ApiKeyAuth
func ThumbsUpStatus(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname = invoker.UName
		id    = c.Param("id")
	)
	ep, err := em.NewEventParticipate(id, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	has := ep.HasPtp("ThumbsUp", uname)
	ptps, err := ep.Ptps("ThumbsUp")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, struct {
		ThumbsUp bool
		Count    int
	}{
		has, len(ptps),
	})
}
