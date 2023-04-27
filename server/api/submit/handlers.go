package submit

import (
	"encoding/json"
	"net/http"

	em "github.com/digisan/event-mgr"
	. "github.com/digisan/go-generics/v2"
	fd "github.com/digisan/gotk/file-dir"
	lk "github.com/digisan/logkit"
	u "github.com/digisan/user-mgr/user"
	"github.com/labstack/echo/v4"
	. "github.com/wismed-web/vhub-api/server/api/definition"
)

// *** after implementing, register with path in 'submit.go' *** //

// @Title Post template
// @Summary get Post template for submission reference.
// @Description
// @Tags    Submit
// @Accept  json
// @Produce json
// @Success 200 "OK - get template successfully"
// @Router /api/submit/template [get]
// @Security ApiKeyAuth
func Template(c echo.Context) error {
	return c.JSON(http.StatusOK, Post{
		Topic:       "Post Topic",
		Type:        "Post or Comment",
		Category:    "Category for this Post. If multiple, Separated by Semicolon",
		Keywords:    "Keywords for this Post. If multiple, Separated by Semicolon",
		ContentHTML: "Post Content, including format feature",
		ContentTEXT: "Post Content, plain text",
	})
}

// @Title Submit a Post
// @Summary submit a Post by filling its template.
// @Description
// @Tags    Submit
// @Accept  json
// @Produce json
// @Param   data      body  string true  "filled Post template json file"
// @Param   followee  query string false "followee Post ID (empty when submitting a new post)"
// @Success 200 "OK - submit successfully"
// @Failure 400 "Fail - incorrect Post format"
// @Failure 500 "Fail - internal error"
// @Router /api/submit/upload [post]
// @Security ApiKeyAuth
func Upload(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname = invoker.UName
		flwee = c.QueryParam("followee")
	)

	P := new(Post)
	if err := c.Bind(P); err != nil {
		lk.Warn("incorrect Uploaded Post Format: %v", err.Error())
		return c.String(http.StatusBadRequest, "incorrect Post format: "+err.Error())
	}
	lk.Log("Uploading ---> [%s] --- %v", uname, P)

	// validating...
	//
	if len(P.Topic) == 0 {
		return c.String(http.StatusBadRequest, "Post Title CANNOT be Empty")
	}
	if len(P.ContentTEXT) == 0 {
		return c.String(http.StatusBadRequest, "Post Content CANNOT be Empty")
	}

	// set P Type
	//
	P.Type = IF(len(flwee) == 0, "P", "C") // P: Post; C: Comment

	// save P as JSON for event
	//
	data, err := json.Marshal(P)
	if err != nil {
		lk.Warn("%v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	evt := em.NewEvent("", uname, "Submit", ConstBytesToStr(data), flwee)
	// lk.Log("event id: [%s]", evt.ID)
	if len(evt.ID) > 0 {
		if err = em.AddEvent(evt); err != nil {
			lk.Warn("%v", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}

		// DEBUG
		fd.MustAppendFile("./debug.txt", []byte(evt.ID), true)

		// FOLLOWING...
		if len(flwee) > 0 {
			ef, err := em.FetchFollow(flwee)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			if ef == nil {
				if ef, err = em.NewEventFollow(flwee, true); err != nil {
					return c.String(http.StatusInternalServerError, err.Error())
				}
			}
			if err := ef.AddFollower(evt.ID); err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
		}
	}
	return c.JSON(http.StatusOK, evt.ID)
}
