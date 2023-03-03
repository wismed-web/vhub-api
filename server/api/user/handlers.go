package user

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	fm "github.com/digisan/file-mgr"
	. "github.com/digisan/go-generics/v2"
	lk "github.com/digisan/logkit"
	si "github.com/digisan/user-mgr/sign-in"
	so "github.com/digisan/user-mgr/sign-out"
	su "github.com/digisan/user-mgr/sign-up"
	u "github.com/digisan/user-mgr/user"
	"github.com/labstack/echo/v4"
	ext "github.com/wismed-web/vhub-api/server/api/user/external"
)

// *** after implementing, register with path in 'user.go' *** //

var (
	UserCache    = &sync.Map{} // map[string]*u.User, *** record logged-in user ***
	MapUserSpace = &sync.Map{} // map[string]*fm.UserSpace, *** record logged-in user space ***
)

// @Title   get password rule
// @Summary get password rule for sign up
// @Description
// @Tags    User
// @Accept  json
// @Produce json
// @Success 200 "OK - got password rule"
// @Router /api/user/pub/pwdrule [get]
func PwdRule(c echo.Context) error {
	lk.Log("Enter: GetPwdRule")
	return c.JSON(http.StatusOK, su.PwdRule())
}

// @Title   register a new user
// @Summary sign up action, send user's basic info for registry
// @Description
// @Tags    User
// @Accept  multipart/form-data
// @Produce json
// @Param   uname formData string true "unique user name"
// @Param   email formData string true "user's email" Format(email)
// @Param   pwd   formData string true "user's password"
// @Success 200 "OK - then waiting for verification code"
// @Failure 400 "Fail - invalid registry fields"
// @Failure 500 "Fail - internal error"
// @Router /api/user/pub/sign-up [post]
func NewUser(c echo.Context) error {

	lk.Debug("[%v] [%v] [%v]", c.FormValue("uname"), c.FormValue("email"), c.FormValue("pwd"))

	user := &u.User{
		Core: u.Core{
			UName:    c.FormValue("uname"),
			Email:    c.FormValue("email"),
			Password: c.FormValue("pwd"),
		},
		Profile: u.Profile{
			Name:           "",
			Phone:          "",
			Country:        "",
			City:           "",
			Addr:           "",
			PersonalIDType: "",
			PersonalID:     "",
			Gender:         "",
			DOB:            "",
			Position:       "",
			Title:          "",
			Employer:       "",
			Bio:            "",
			AvatarType:     "",
			Avatar:         []byte{},
		},
		Admin: u.Admin{
			RegTime:   time.Now().Truncate(time.Second),
			Active:    true,
			Certified: false,
			Official:  false,
			SysRole:   "",
			MemLevel:  0,
			MemExpire: time.Time{},
			Tags:      "",
		},
	}

	// su.SetValidator(map[string]func(string) bool{ })

	lk.Log("%v", user)

	if err := su.ChkInput(user); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if err := su.ChkEmail(user); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, "waiting verification code in your email")

	///////////////////////////////////////////////
	// simple sing up, ignore email verification //
	///////////////////////////////////////////////

	// // store into db
	// if err := su.Store(user); err != nil {
	// 	return c.String(http.StatusInternalServerError, err.Error())
	// }
	// // sign-up ok calling...
	// {
	// }
	// return c.JSON(http.StatusOK, "registered successfully")
}

// @Title   verify new user's email
// @Summary sign up action, step 2. send back email verification code
// @Description
// @Tags    User
// @Accept  multipart/form-data
// @Produce json
// @Param   uname formData string true "unique user name"
// @Param   code  formData string true "verification code (in user's email)"
// @Success 200 "OK - sign-up successfully"
// @Failure 400 "Fail - incorrect verification code"
// @Failure 500 "Fail - internal error"
// @Router /api/user/pub/verify-email [post]
func VerifyEmail(c echo.Context) error {

	var (
		uname = c.FormValue("uname")
		code  = c.FormValue("code")
	)

	user, err := su.VerifyCode(uname, code)
	if err != nil || user == nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// double check before storing
	if err := su.ChkInput(user); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// if user is admin, then
	if In(user.UName, admins...) {
		user.SysRole = "admin"
	}

	// store into db
	if err := su.Store(user); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// sign-up ok calling...
	{
	}

	return c.JSON(http.StatusOK, "sign up successfully")
}

// @Title sign in
// @Summary sign in action. if ok, got token
// @Description
// @Tags    User
// @Accept  multipart/form-data
// @Produce json
// @Param   uname formData string true "user name or email"
// @Param   pwd   formData string true "password" Format(password)
// @Success 200 "OK - sign-in successfully"
// @Failure 400 "Fail - incorrect password"
// @Failure 500 "Fail - internal error"
// @Router /api/user/pub/sign-in [post]
func LogIn(c echo.Context) error {

	var (
		uname       = c.FormValue("uname")
		pwd         = c.FormValue("pwd")
		email       = c.FormValue("uname")
		loginFailed = false
	)

	lk.Log("login: [%v] [%v]", uname, pwd)

	user := &u.User{
		Core: u.Core{
			UName:    uname,
			Password: pwd,
			Email:    email,
		},
		Profile: u.Profile{},
		Admin:   u.Admin{},
	}

	defer func() {
		if loginFailed {
			si.CheckFrequentlyAccess(uname, 10, 3)
		} else {
			si.RemoveFrequentlyAccessRecord(uname, 1*time.Millisecond)
		}
	}()

	if si.IsFrequentlyAccess(uname) {
		loginFailed = true
		si.RemoveFrequentlyAccessRecord(uname, 15*time.Second)
		return c.String(http.StatusBadRequest, "Failed frequently, please try to Login later")
	}

AGAIN:

	if err := si.UserStatusIssue(user); err != nil {

		///////////////////////////////////////
		// external user checking
		{
			// try V-HUB existing user check. if external user already exists, wrap user & login again
			if u := ext.ValidateSavedExtUser(uname, pwd); u != nil {
				user = u
				goto AGAIN
			}

			// external V-HUB check via remote api
			if ok, err := ext.ExtUserLoginCheck(uname, pwd); err == nil && ok {
				// if CAN login V-HUB, but doesn't exist, now create a new external user, its uname is like "13888888888@@V"
				u, err := ext.CreateExtUser(uname, pwd)
				if err != nil {
					return c.String(http.StatusInternalServerError, "ERR: creating external user, "+err.Error())
				}
				user = u
				goto AGAIN
			}
		}
		///////////////////////////////////////
		loginFailed = true
		return c.String(http.StatusBadRequest, err.Error())
	}

	if !si.PwdOK(user) { // if successful, user updated.
		loginFailed = true
		return c.String(http.StatusBadRequest, "incorrect password")
	}

	// fmt.Println(user)

	// now, user is real user in db
	defer lk.FailOnErr("%v", si.Trail(user.UName)) // Refresh Online Users, here UName is real

	// log in ok calling...
	{
		{
			// MUST DO `fm.InitFileMgr` in advance in elsewhere
			us, err := fm.UseUser(user.UName)
			if err != nil || us == nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}
			MapUserSpace.Store(user.UName, us)
		}
	}

	defer func() { UserCache.Store(user.UName, user) }() // save current user for other usage

	claims := u.MakeClaims(user)
	token := u.GenerateToken(claims)
	return c.JSON(http.StatusOK, echo.Map{
		"token": token,
		"auth":  "Bearer " + token,
	})
}

// @Title   trail
// @Summary trail alive user.
// @Description
// @Tags    User
// @Accept  json
// @Produce json
// @Success 200 "OK - trail successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/user/auth/trail [patch]
// @Security ApiKeyAuth
func Trail(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		lk.Warn("%v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	uname := invoker.UName
	if err := si.Trail(uname); err != nil {
		lk.Warn("%v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, fmt.Sprintf("[%s] trails successfully", uname))
}

// @Title   sign out
// @Summary sign out action.
// @Description
// @Tags    User
// @Accept  json
// @Produce json
// @Success 200 "OK - sign-out successfully"
// @Failure 500 "Fail - internal error"
// @Router /api/user/auth/sign-out [put]
// @Security ApiKeyAuth
func LogOut(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		lk.Warn("%v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	defer invoker.DeleteToken() // only in SignOut calling DeleteToken()

	uname := invoker.UName

	// remove user by 'uname'
	defer UserCache.Delete(uname)

	if err := so.Logout(uname); err != nil {
		lk.Warn("%v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, fmt.Sprintf("[%s] sign-out successfully", uname))
}

// @Title   get uname
// @Summary get uname
// @Description
// @Tags    User
// @Accept  json
// @Produce json
// @Success 200 "OK - got uname"
// @Router /api/user/auth/uname [get]
// @Security ApiKeyAuth
func GetUname(c echo.Context) error {

	lk.Log("Enter: GetUname")

	user, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// lk.Debug(uname)
	return c.JSON(http.StatusOK, user.UName)
}
