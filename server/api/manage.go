package api

import (
	"github.com/labstack/echo/v4"
	"github.com/wismed-web/vhub-api/server/api/manage"
)

// register to main echo Group

// "/api/manage"
func ManageHandler(e *echo.Group) {

	var mGET = map[string]echo.HandlerFunc{
		"/own/id":               manage.OwnPosts,
		"/bookmark/status/:id":  manage.BookmarkStatus,
		"/bookmark/marked":      manage.BookmarkedPosts,
		"/follower/id":          manage.Followers,
		"/thumbs-up/status/:id": manage.ThumbsUpStatus,
	}
	var mPOST = map[string]echo.HandlerFunc{}
	var mPUT = map[string]echo.HandlerFunc{}
	var mDELETE = map[string]echo.HandlerFunc{
		"/delete/one": manage.DelOne,
		"/erase/one":  manage.EraseOne,
	}
	var mPATCH = map[string]echo.HandlerFunc{
		"/bookmark/:id":  manage.Bookmark,
		"/thumbs-up/:id": manage.ThumbsUp,
	}

	// ------------------------------------------------------- //

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	mRegAPIs := map[string]map[string]echo.HandlerFunc{
		"GET":    mGET,
		"POST":   mPOST,
		"PUT":    mPUT,
		"DELETE": mDELETE,
		"PATCH":  mPATCH,
		// others...
	}

	mRegMethod := map[string]func(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route{
		"GET":    e.GET,
		"POST":   e.POST,
		"PUT":    e.PUT,
		"DELETE": e.DELETE,
		"PATCH":  e.PATCH,
		// others...
	}

	for _, m := range methods {
		mAPI, method := mRegAPIs[m], mRegMethod[m]
		for path, handler := range mAPI {
			if handler == nil {
				continue
			}
			method(path, handler)
		}
	}
}
