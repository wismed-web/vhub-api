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

// @Title   delete Post ID
// @Summary delete one Post ID from timeline.
// @Description
// @Tags    Manage
// @Accept  json
// @Produce json
// @Param   id   path string true "Post ID for deleting"
// @Success 200 "OK - delete successfully"
// @Failure 400 "Fail - incorrect path param 'id'"
// @Failure 405 "Fail - invoker's role is NOT in permit group"
// @Failure 500 "Fail - internal error"
// @Router /api/manage/debug/delete/{id} [delete]
// @Security ApiKeyAuth
func DelGlobalPostID(c echo.Context) error {

	invoker, err := u.ToActiveFullUser(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if NotIn(invoker.SysRole, "system") {
		return c.String(http.StatusMethodNotAllowed, "only system level users can do Debug DELETE")
	}

	var (
		id = c.Param("id")
	)
	if len(id) == 0 {
		return c.String(http.StatusBadRequest, "post 'id' cannot be empty")
	}
	deleted, err := em.DelGlobalEventID(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, IF(len(deleted) == 1, fmt.Sprintf("<%s> is deleted", id), fmt.Sprintf("<%s> is not deleted successfully", id)))
}

// @Title   delete Post
// @Summary delete one Post content.
// @Description
// @Tags    Manage
// @Accept  json
// @Produce json
// @Param   id   path string true "Post ID for deleting"
// @Success 200 "OK - delete successfully"
// @Failure 400 "Fail - incorrect path param 'id'"
// @Failure 405 "Fail - invoker's role is NOT in permit group"
// @Failure 500 "Fail - internal error"
// @Router /api/manage/delete/{id} [delete]
// @Security ApiKeyAuth
func DelOnePost(c echo.Context) error {

	invoker, err := u.ToActiveFullUser(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if NotIn(invoker.SysRole, "admin", "system") {
		return c.String(http.StatusMethodNotAllowed, "only admin or system level users can do DELETE")
	}

	var (
		id = c.Param("id")
	)
	if len(id) == 0 {
		return c.String(http.StatusBadRequest, "post 'id' cannot be empty")
	}
	deleted, err := em.DelEvent(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, IF(len(deleted) == 1, fmt.Sprintf("<%s> is deleted", id), fmt.Sprintf("<%s> is not deleted successfully", id)))
}

// @Title erase one Post content
// @Summary erase one Post content permanently.
// @Description
// @Tags    Manage
// @Accept  json
// @Produce json
// @Param   id   path string true "Post ID for erasing"
// @Success 200 "OK - erase successfully"
// @Failure 400 "Fail - incorrect path param 'id'"
// @Failure 405 "Fail - invoker's role is NOT in permit group"
// @Failure 500 "Fail - internal error"
// @Router /api/manage/erase/{id} [delete]
// @Security ApiKeyAuth
func EraseOnePost(c echo.Context) error {

	invoker, err := u.ToActiveFullUser(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if NotIn(invoker.SysRole, "admin", "system") {
		return c.String(http.StatusMethodNotAllowed, "only admin or system level users can do DELETE")
	}

	var (
		id = c.Param("id")
	)
	if len(id) == 0 {
		return c.String(http.StatusBadRequest, "post 'id' cannot be empty")
	}
	erased, err := em.EraseEvent(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, IF(len(erased) == 1, fmt.Sprintf("<%s> is erased permanently", id), fmt.Sprintf("<%s> is not erased successfully", id)))
}

// @Title get own Post id group in a specific period
// @Summary get own Post id group in one specific time period.
// @Description
// @Tags    Manage
// @Accept  json
// @Produce json
// @Param   period query string false "time period for query, format is 'yyyymm', e.g. '202206'. if missing, current yyyymm applies"
// @Success 200 "OK - get successfully"
// @Failure 400 "Fail - incorrect query param type"
// @Failure 500 "Fail - internal error"
// @Router /api/manage/own [get]
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
