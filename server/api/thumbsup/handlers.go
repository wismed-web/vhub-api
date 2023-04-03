package thumbsup

import (
	"net/http"

	em "github.com/digisan/event-mgr"
	u "github.com/digisan/user-mgr/user"
	"github.com/labstack/echo/v4"
)

// @Title toggle a thumbs-up
// @Summary toggle a personal thumbs-up for a post.
// @Description
// @Tags    ThumbsUp
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for adding or removing thumbs-up"
// @Success 200 "OK - added or removed thumb successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/thumbs-up/toggle/{id} [patch]
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
	has, err := ep.ToggleParticipant("ThumbsUp", uname)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	participants, err := ep.Participants("ThumbsUp")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, struct {
		ThumbsUp bool
		Count    int
	}{
		has,
		len(participants),
	})
}

// @Title current user's thumbs-up status
// @Summary get current login user's thumbs-up status for a post.
// @Description
// @Tags    ThumbsUp
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for checking thumbs-up status"
// @Success 200 "OK - get thumbs-up status successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/thumbs-up/status/{id} [get]
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
	ep, err := em.NewEventParticipate(id, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	has := ep.HasParticipant("ThumbsUp", uname)
	ptps, err := ep.Participants("ThumbsUp")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, struct {
		ThumbsUp bool
		Count    int
	}{
		has,
		len(ptps),
	})
}
