package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	. "github.com/digisan/go-generics/v2"
	gm "github.com/digisan/go-mail"
	lk "github.com/digisan/logkit"
	u "github.com/digisan/user-mgr/user"
	"github.com/labstack/echo/v4"
)

// @Title   send email
// @Summary send email by 3rd service
// @Description
// @Tags    Admin
// @Accept  multipart/form-data
// @Produce json
// @Param   unames  formData string true "unique user names, separator is ',' "
// @Param   subject formData string true "subject for email"
// @Param	body    formData string true "body for email"
// @Success 200 "OK - list successfully"
// @Failure 401 "Fail - unauthorized error"
// @Failure 403 "Fail - forbidden error"
// @Failure 500 "Fail - internal error"
// @Router /api/admin/email [post]
// @Security ApiKeyAuth
func SendEmail(c echo.Context) error {

	lk.Warn("Enter: SendEmail")

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if _, ok, _ := u.LoadActiveUser(invoker.UName); !ok {
		return c.String(http.StatusForbidden, fmt.Sprintf("invalid invoker status@[%s], dormant?", invoker.UName))
	}

	/////////////////////////////////////

	const (
		sep = "," // separator for unames
	)

	var (
		unames  = c.FormValue("unames")  // recipients, separator is ','
		subject = c.FormValue("subject") // email title
		body    = c.FormValue("body")    // email content
	)

	type retType struct {
		OK     bool
		Sent   []string
		Failed []string
		Err    []error
	}
	ret := []retType{}

	for _, uname := range strings.Split(unames, sep) {
		lk.Log("[%v] [%v] [%v]", uname, subject, body)

		user, ok, err := u.LoadUser(uname, true)
		switch {
		case err != nil:
			return c.String(http.StatusInternalServerError, err.Error())
		case !ok:
			return c.String(http.StatusBadRequest, fmt.Sprintf("[%s] doesn't exist", uname))
		}

		ok, sent, failed, errs := gm.SendMail(subject, body, user.Email)
		ret = append(ret, retType{
			OK:     ok,
			Sent:   sent,
			Failed: failed,
			Err:    errs,
		})
	}

	return c.JSON(http.StatusOK, ret)
}

// @Title   remove user
// @Summary remove an user by its uname
// @Description
// @Tags    Admin
// @Accept  json
// @Produce json
// @Param   uname path string true "uname of the user to be removed"
// @Success 200 "OK - remove successfully"
// @Failure 401 "Fail - unauthorized error"
// @Failure 500 "Fail - internal error"
// @Router /api/admin/user/remove/{uname} [delete]
// @Security ApiKeyAuth
func RemoveUser(c echo.Context) error {

	lk.Log("Enter: RemoveUser")

	admin, err := u.ToActiveFullUser(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if admin.SysRole != "admin" {
		return c.String(http.StatusUnauthorized, fmt.Sprintf("only admin can do removing action"))
	}

	//////////////////////////////////////////////

	var (
		uname = c.Param("uname")
	)
	if err := u.RemoveUser(uname, true); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, fmt.Sprintf("%s is removed successfully", uname))
}

// @Title   list users' info
// @Summary list users' info
// @Description
// @Tags    Admin
// @Accept  json
// @Produce json
// @Param   uname  query string false "user filter with uname wildcard(*)"
// @Param   name   query string false "user filter with name wildcard(*)"
// @Param   active query string false "user filter with active status"
// @Param   fields path  string false "which user's fields (sep by ',') want to list. if empty, return all fields"
// @Success 200 "OK - list successfully"
// @Failure 401 "Fail - unauthorized error"
// @Failure 403 "Fail - forbidden error"
// @Failure 500 "Fail - internal error"
// @Router /api/admin/user/list/{fields} [get]
// @Security ApiKeyAuth
func ListUser(c echo.Context) error {

	lk.Log("Enter: ListUser")

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if _, ok, _ := u.LoadActiveUser(invoker.UName); !ok {
		return c.String(http.StatusForbidden, fmt.Sprintf("invalid invoker status@[%s], dormant?", invoker.UName))
	}

	//////////////////////////////////////////////

	var (
		active = c.QueryParam("active")
		wUname = c.QueryParam("uname")
		wName  = c.QueryParam("name")
		rUname = wc2re(wUname)
		rName  = wc2re(wName)
		fields = c.Param("fields")
	)

	users, err := u.ListUser(func(u *u.User) bool {
		switch {
		case len(wUname) > 0 && !rUname.MatchString(u.UName):
			return false
		case len(wName) > 0 && !rName.MatchString(u.Name):
			return false
		case len(active) > 0:
			if bActive, err := strconv.ParseBool(active); err == nil {
				return bActive == u.Active
			}
			return false
		default:
			return true
		}
	})

	for _, user := range users {
		user.Password = strings.Repeat("*", len(user.Password))
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// convert url special symbol string to normal characters
	if fields, err = url.QueryUnescape(fields); err != nil {
		c.String(http.StatusBadRequest, "'fields' is invalid")
	}
	// *** if 'fields' is not provided, swagger "Try" put it value as string "{fields}" or "undefined" ***
	if In(fields, "{fields}", "undefined", "") {
		return c.JSON(http.StatusOK, users)
	}

	// lk.Debug("%v", fields)

	fieldsUser := []string{}
	for _, field := range strings.Split(fields, ",") {
		fieldsUser = AppendIf(In(field, "uname", "Uname", "ID", "Id", "id"), fieldsUser, "UName")
		fieldsUser = AppendIf(In(field, "email", "Email"), fieldsUser, "Email")
		fieldsUser = AppendIf(In(field, "name", "Name"), fieldsUser, "Name")
	}
	rt := FilterMap(users, nil, func(i int, e *u.User) any {
		v, err := PartialAsMap(e, fieldsUser...)
		lk.WarnOnErr("%v", err)
		return v
	})
	return c.JSON(http.StatusOK, rt)
}

// @Title   list online users
// @Summary get all online users
// @Description
// @Tags    Admin
// @Accept  json
// @Produce json
// @Param   uname query string false "user filter with uname wildcard(*)"
// @Success 200 "OK - list successfully"
// @Failure 401 "Fail - unauthorized error"
// @Failure 403 "Fail - forbidden error"
// @Failure 500 "Fail - internal error"
// @Router /api/admin/user/online [get]
// @Security ApiKeyAuth
func ListOnlineUser(c echo.Context) error {

	lk.Log("Enter: ListOnlineUser")

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if _, ok, _ := u.LoadActiveUser(invoker.UName); !ok {
		return c.String(http.StatusForbidden, fmt.Sprintf("invalid invoker status@[%s], dormant?", invoker.UName))
	}

	//////////////////////////////////////////////

	var (
		wUname = c.QueryParam("uname")
		rUname = wc2re(wUname)
	)

	online, err := u.OnlineUsers()
	// for _, user := range online {
	// 	fmt.Println(user)
	// }
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	FilterFast(&online, func(i int, e *u.UserOnline) bool {
		if len(wUname) > 0 && !rUname.MatchString(e.Uname) {
			return false
		}
		return true
	})

	return c.JSON(http.StatusOK, online)
}

// @Title   update user's info
// @Summary update user's info by fields & its values
// @Description
// @Tags    Admin
// @Accept  multipart/form-data
// @Produce json
// @Param   uname  formData string true "unique user name want to be updated"
// @Param   fields path     string true "which user struct fields (sep by ',') want to be updated. (fields must be IDENTICAL TO STRUCT FIELDS !!!)"
// @Success 200 "OK - list successfully"
// @Failure 400 "Fail - bad request error"
// @Failure 401 "Fail - unauthorized error"
// @Failure 500 "Fail - internal error"
// @Router /api/admin/user/update/{fields} [put]
// @Security ApiKeyAuth
func UpdateUser(c echo.Context) error {

	lk.Log("Enter: UpdateUser")

	if _, err := u.ToActiveFullUser(c); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	//////////////////////////////////////////////

	var (
		uname  = c.FormValue("uname") // ***
		fields = c.Param("fields")    // sep by ','
	)

	user, ok, err := u.LoadAnyUser(uname)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if !ok {
		return c.String(http.StatusBadRequest, fmt.Sprintf("'%s' is not existing, unable to update", uname))
	}

	if fields, err = url.QueryUnescape(fields); err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	if In(fields, "{fields}", "undefined", "") {
		return c.String(http.StatusBadRequest, "updating 'fields' must be provided")
	}

	// url 'field' must be exportable & identical to struct definition !!!
	for _, field := range strings.Split(fields, ",") {
		// set struct, field name must be exportable !!!
		val := c.FormValue(field) // *** c.FormValue here, Field Names Must be exportable & identical to url Field Names ***
		// lk.Debug("%v", val)
		if err = SetField(user, field, val); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
	}

	// update db
	if err = u.UpdateUser(user); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	// lk.Debug("%+v", *user)

	return c.JSON(http.StatusOK, fmt.Sprintf("'%v' has been updated", user))
}

// @Title    get any user's avatar
// @Summary  get any user's avatar src as base64
// @Description
// @Tags     Admin
// @Accept   json
// @Produce  json
// @Param    uname query string true "user registered unique name"
// @Success  200 "OK - get avatar src base64"
// @Failure  400 "Fail - cannot find user via given uname"
// @Failure  404 "Fail - avatar is empty"
// @Failure  500 "Fail - internal error"
// @Router   /api/admin/user/avatar [get]
// @Security ApiKeyAuth
func GetAvatar(c echo.Context) error {
	var (
		uname = c.QueryParam("uname")
	)
	if len(uname) == 0 {
		return c.String(http.StatusBadRequest, "[uname] cannot be empty")
	}
	user, ok, err := u.LoadAnyUser(uname)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if !ok {
		return c.String(http.StatusBadRequest, fmt.Sprintf("cannot find user [%v]", uname))
	}
	// if fetch extra fields from a user, must make sure it is full fields
	b64, aType := user.AvatarBase64(false)
	if len(b64) == 0 || len(aType) == 0 {
		return c.String(http.StatusNotFound, "avatar is empty")
	}
	return c.JSON(http.StatusOK, struct {
		Src string `json:"src"`
	}{
		Src: fmt.Sprintf("data:%s;base64,%s", aType, b64),
	})
}

// @Title    get user info
// @Summary  get any user info
// @Description
// @Tags     Admin
// @Accept   json
// @Produce  json
// @Param    uname query string true "user registered unique name"
// @Success  200   "OK - get info"
// @Failure  400   "Fail - cannot find user via given uname"
// @Failure  500   "Fail - internal error"
// @Router   /api/admin/user/info [get]
// @Security ApiKeyAuth
func GetUserInfo(c echo.Context) error {
	var (
		uname = c.QueryParam("uname")
	)
	if len(uname) == 0 {
		return c.String(http.StatusBadRequest, "[uname] cannot be empty")
	}
	user, ok, err := u.LoadAnyUser(uname)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if !ok {
		return c.String(http.StatusBadRequest, fmt.Sprintf("cannot find user [%v]", uname))
	}
	return c.JSON(http.StatusOK, *user)
}

// @Title    get user fields value
// @Summary  get any user some fields value
// @Description
// @Tags     Admin
// @Accept   json
// @Produce  json
// @Param    uname  query string true "user registered unique name"
// @Param    fields path  string true "which user struct fields (sep by ',') want to be fetched. (fields must be IDENTICAL TO STRUCT FIELDS !!!)"
// @Success  200    "OK - get info"
// @Failure  400    "Fail - cannot find user via given uname"
// @Failure  500    "Fail - internal error"
// @Router   /api/admin/user/field-value/{fields} [get]
// @Security ApiKeyAuth
func GetUserFieldValue(c echo.Context) error {
	var (
		uname  = c.QueryParam("uname") // ***
		fields = c.Param("fields")     // sep by ','
		err    error
	)
	if fields, err = url.QueryUnescape(fields); err != nil {
		c.String(http.StatusBadRequest, err.Error())
	}
	if len(uname) == 0 {
		return c.String(http.StatusBadRequest, "[uname] cannot be empty")
	}
	user, ok, err := u.LoadAnyUser(uname)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if !ok {
		return c.String(http.StatusBadRequest, fmt.Sprintf("cannot find user [%v]", uname))
	}
	rt := make(map[string]any)
	// url 'field' must be exportable & identical to struct definition !!!
	for _, field := range strings.Split(fields, ",") {
		val, err := FieldValue(*user, field)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("cannot fetch field [%v] from [%v]", field, uname))
		}
		rt[field] = val
	}
	return c.JSON(http.StatusOK, rt)
}

// @Title   list user's action record
// @Summary list user's action record
// @Description
// @Tags    Admin
// @Accept  json
// @Produce json
// @Param   uname  query string true "user registered unique name"
// @Param   action path  string true "which action type [submit, approve, subscribe] record want to list"
// @Success 200 "OK - list successfully"
// @Failure 401 "Fail - unauthorized error"
// @Failure 403 "Fail - forbidden error"
// @Failure 500 "Fail - internal error"
// @Router /api/admin/user/action-list/{action} [get]
// @Security ApiKeyAuth
// func ListUserAction(c echo.Context) error {

// 	var (
// 		userTkn = c.Get("user").(*jwt.Token)
// 		claims  = userTkn.Claims.(*u.UserClaims)
// 	)

// 	user, ok, err := u.LoadActiveUser(claims.UName)
// 	switch {
// 	case err != nil:
// 		return c.String(http.StatusInternalServerError, err.Error())
// 	case !ok:
// 		return c.String(http.StatusForbidden, fmt.Sprintf("invalid user status@[%s], dormant?", user.UName))
// 	}

// 	// --- //

// 	var (
// 		uname  = c.QueryParam("uname") // other uname
// 		action = c.Param("action")     // action type: submit, approve, subscribe
// 	)

// 	ls, err := db.ListActionRecord(uname, db.DbColType(action))
// 	if err != nil {
// 		return c.String(http.StatusInternalServerError, err.Error())
// 	}
// 	return c.JSON(http.StatusOK, ls)
// }
