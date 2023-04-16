package retrieve

import (
	"encoding/json"
	"fmt"
	"net/http"

	em "github.com/digisan/event-mgr"
	. "github.com/digisan/go-generics/v2"
	lk "github.com/digisan/logkit"
	u "github.com/digisan/user-mgr/user"
	"github.com/labstack/echo/v4"
	. "github.com/wismed-web/vhub-api/server/api/definition"
)

// @Title   get a batch of Post id
// @Summary get a batch of Post id group.
// @Description
// @Tags    Retrieve
// @Accept  json
// @Produce json
// @Param   by    query string true "'time'"
// @Param   value query string true "recent [value] minutes for time"
// @Success 200 "OK - get successfully"
// @Failure 400 "Fail - incorrect query param type"
// @Failure 404 "Fail - not found"
// @Failure 500 "Fail - internal error"
// @Router /api/retrieve/batch-id [get]
// @Security ApiKeyAuth
func BatchID(c echo.Context) error {
	var (
		by    = c.QueryParam("by")
		value = c.QueryParam("value")
	)
	_, ok := AnyTryToType[int](value)
	if !ok {
		return c.String(http.StatusBadRequest, "'value' must be a valid number for both time(minutes)")
	}
	switch by {
	case "time", "tm":
		ids, err := em.FetchEvtIDByTm(value + "m")
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, ids)
	// case "count", "cnt":
	// 	ids, err := em.FetchEvtIDsByCnt(int(n), "30m", "1h", "2h", "6h", "12h", "24h")
	// 	if err != nil {
	// 		return c.String(http.StatusInternalServerError, err.Error())
	// 	}
	// 	if len(ids) < n {
	// 		idAll, err := em.FetchEvtIDs(nil)
	// 		if err != nil {
	// 			return c.String(http.StatusInternalServerError, err.Error())
	// 		}
	// 		if len(idAll) > n {
	// 			ids = idAll[:n]
	// 		}
	// 	}
	// 	return c.JSON(http.StatusOK, ids)
	default:
		return c.String(http.StatusBadRequest, "'by' must be [time]")
	}
}

// @Title   get all Post id
// @Summary get all Post id group.
// @Description
// @Tags    Retrieve
// @Accept  json
// @Produce json
// @Success 200 "OK - get successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/retrieve/all-id [get]
// @Security ApiKeyAuth
func AllID(c echo.Context) error {
	ids, err := em.FetchEvtID(nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	// lk.Log("IdAll ---> %d : %v", len(ids), ids)
	return c.JSON(http.StatusOK, ids)
}

// @Title   get all Span
// @Summary get all Span group.
// @Description
// @Tags    Retrieve
// @Accept  json
// @Produce json
// @Success 200 "OK - get successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/retrieve/debug/all-span [get]
// @Security ApiKeyAuth
func AllSpan(c echo.Context) error {

	invoker, err := u.ToActiveFullUser(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if NotIn(invoker.SysRole, "system") {
		return c.String(http.StatusMethodNotAllowed, "only system level users can do Debug AllSpan")
	}

	spans, err := em.FetchSpan(nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	// lk.Log("All Spans ---> %d : %v", len(spans), spans)
	return c.JSON(http.StatusOK, spans)
}

// @Title get one Post content
// @Summary get one Post content.
// @Description
// @Tags    Retrieve
// @Accept  json
// @Produce json
// @Param   id     query string  true "Post ID for its content"
// @Success 200 "OK - get Post event successfully"
// @Failure 400 "Fail - incorrect query param id"
// @Failure 404 "Fail - not found"
// @Failure 500 "Fail - internal error"
// @Router /api/retrieve/post [get]
// @Security ApiKeyAuth
func OnePost(c echo.Context) error {
	var (
		id = c.QueryParam("id")
	)
	lk.Log("Into GetOne, event id is %v", id)
	if len(id) == 0 {
		return c.String(http.StatusBadRequest, "'id' cannot be empty")
	}
	event, err := em.FetchEvent(true, id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if event == nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("Post not found @%s", id))
	}
	if len(event.RawJSON) == 0 {
		return c.JSON(http.StatusOK, fmt.Sprintf("Post is empty @%s", id))
	}
	////////////////////////////////////
	// set up event content, i.e. Post
	P := &Post{}
	if err := json.Unmarshal([]byte(event.RawJSON), P); err != nil {
		lk.Warn("Unmarshal Post Error, event is %v", event)
		return c.String(http.StatusInternalServerError, "convert RawJSON to [Post] Unmarshal error")
	}
	return c.JSON(http.StatusOK, event)
}
