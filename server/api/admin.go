package api

import (
	"github.com/labstack/echo/v4"
	ad "github.com/wismed-web/vhub-api/server/api/admin"
)

// register to main echo Group

// /api/admin
func AdminHandler(r *echo.Group) {

	var mGET = map[string]echo.HandlerFunc{
		"/user/list/:fields":        ad.ListUser,
		"/user/online":              ad.ListOnlineUser,
		"/user/avatar":              ad.GetAvatar,
		"/user/info":                ad.GetUserInfo,
		"/user/field-value/:fields": ad.GetUserFieldValue,
		// "/user/action-list/:action": ad.ListUserAction,
	}

	var mPOST = map[string]echo.HandlerFunc{
		"/email": ad.SendEmail,
	}

	var mPUT = map[string]echo.HandlerFunc{
		"/user/update/:fields": ad.UpdateUser,
	}

	var mDELETE = map[string]echo.HandlerFunc{
		"/user/remove/:uname": ad.RemoveUser,
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
		"GET":    r.GET,
		"POST":   r.POST,
		"PUT":    r.PUT,
		"DELETE": r.DELETE,
		"PATCH":  r.PATCH,
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
