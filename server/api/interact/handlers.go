package interact

import (
	"net/http"

	em "github.com/digisan/event-mgr"
	. "github.com/digisan/go-generics/v2"
	u "github.com/digisan/user-mgr/user"
	"github.com/labstack/echo/v4"
)

// @Title   toggle an action
// @Summary toggle an action like 'thumbs-up', 'like' for a post.
// @Description
// @Tags    Interact
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for adding or removing an action"
// @Success 200 "OK - added or removed one action successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/interact/{action}/toggle/{id} [patch]
// @Security ApiKeyAuth
func Toggle(c echo.Context) error {
	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	var (
		uname  = invoker.UName
		action = c.Param(("action"))
		id     = c.Param("id")
	)
	if NotIn(action, "ThumbsUp", "Like") {
		return c.String(http.StatusBadRequest, "[action] can only be [ThumbsUp, Like]")
	}
	ep, err := em.NewEventParticipate(id, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	has, err := ep.ToggleParticipant(action, uname)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	participants, err := ep.Participants(action)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, struct {
		Action string
		Status bool
		Count  int
	}{
		action,
		has,
		len(participants),
	})
}

// @Title   current user's one action status
// @Summary get current login user's one action status like 'thumbs-up', 'like' for a post.
// @Description
// @Tags    Interact
// @Accept  json
// @Produce json
// @Param   id path string true "Post ID (event id) for checking an action status"
// @Success 200 "OK - get one action status successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/interact/{action}/status/{id} [get]
// @Security ApiKeyAuth
func Status(c echo.Context) error {
	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	var (
		uname  = invoker.UName
		action = c.Param(("action"))
		id     = c.Param("id")
	)
	if NotIn(action, "ThumbsUp", "Like") {
		return c.String(http.StatusBadRequest, "[action] can only be [ThumbsUp, Like]")
	}
	ep, err := em.NewEventParticipate(id, true)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	has := ep.HasParticipant(action, uname)
	ptps, err := ep.Participants(action)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, struct {
		Action   string
		ThumbsUp bool
		Count    int
	}{
		action,
		has,
		len(ptps),
	})
}
