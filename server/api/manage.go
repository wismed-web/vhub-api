package api

import (
	"github.com/labstack/echo/v4"
	"github.com/wismed-web/vhub-api/server/api/manage"
)

// register to main echo Group

// "/api/manage"
func ManageHandler(e *echo.Group) {

	var mGET = map[string]echo.HandlerFunc{
		"/own": manage.OwnPosts,
	}
	var mPOST = map[string]echo.HandlerFunc{}
	var mPUT = map[string]echo.HandlerFunc{}
	var mDELETE = map[string]echo.HandlerFunc{
		"/delete/:id":       manage.DelOnePost,
		"/erase/:id":        manage.EraseOnePost,
		"/debug/delete/:id": manage.DelGlobalPostID,
	}
	var mPATCH = map[string]echo.HandlerFunc{}

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
