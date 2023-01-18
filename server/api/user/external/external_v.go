package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	lk "github.com/digisan/logkit"
	si "github.com/digisan/user-mgr/sign-in"
	su "github.com/digisan/user-mgr/sign-up"
	u "github.com/digisan/user-mgr/user"
)

const (
	extSep  = "@@" // for creating external user-name e.g. 13888888888@@V
	timeout = 10   // unit second
	// V
	vCode  = "V"          // V Site Code
	vEmail = "wismed.net" // V Site Mail
	// V Site API
	apiVUserExists        = `https://www.scwismed.cn/api/external/checkUser`
	apiVUserExistsToken   = `N39UNuYt3&GEfYS*uxYp7mti1Pv^$ISA`
	apiVUserCanLogin      = `https://www.scwismed.cn/api/external/checkUserIsLogin`
	apiVUserCanLoginToken = `N39UNuYt3&GEfYS*uxYp7mti1Pv^$ISA`
)

func newExtUser(userId, pwd string) *u.User {
	return &u.User{
		Core: u.Core{
			UName:    fmt.Sprintf("%s%s%s", userId, extSep, vCode),
			Email:    fmt.Sprintf("%s@%s", userId, vEmail),
			Password: pwd,
		},
		Profile: u.Profile{
			Name:           userId,
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
			MemLevel:  3,
			MemExpire: time.Time{},
			Tags:      "",
		},
	}
}

func CreateExtUser(userId, pwd string) (*u.User, error) {
	user := newExtUser(userId, pwd)
	return user, su.Store(user)
}

func ValidateSavedExtUser(userId, pwd string) *u.User {
	if user := newExtUser(userId, pwd); si.UserStatusIssue(user) == nil {
		return user
	}
	return nil
}

type ResultExt struct {
	ok  bool
	err error
}

func ExtUserExistsAsync(userId string) chan ResultExt {
	cResult := make(chan ResultExt)
	go func() {
		params := url.Values{}
		params.Add("token", apiVUserExistsToken)
		params.Add("mobile", userId)
		resp, err := http.PostForm(apiVUserExists, params)
		lk.WarnOnErr("%v", err)
		if err != nil {
			cResult <- ResultExt{
				ok:  false,
				err: err,
			}
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		lk.WarnOnErr("%v", err)
		if err != nil {
			cResult <- ResultExt{
				ok:  false,
				err: err,
			}
			return
		}
		mData := make(map[string]any)
		lk.WarnOnErr("%v", json.Unmarshal(body, &mData))

		// fmt.Println(mData)

		if mData["code"].(float64) == 200 {
			if mData["status"] == "SUCCESS" {
				if mData["data"].(map[string]any)["sta"].(float64) == 1 {
					cResult <- ResultExt{
						ok:  true,
						err: nil,
					}
					return
				}
			}
		}
		cResult <- ResultExt{
			ok:  false,
			err: fmt.Errorf("[%s] is not registered in V-Site", userId),
		}
	}()
	return cResult
}

func ExtUserExistsCheck(userId string) (bool, error) {
	select {
	case <-time.After(timeout * time.Second):
		return false, errors.New("timeout")
	case r := <-ExtUserExistsAsync(userId):
		return r.ok, r.err
	}
}

func ExtUserLoginValidateAsync(userId, pwd string) chan ResultExt {
	cResult := make(chan ResultExt)
	go func() {
		params := url.Values{}
		params.Add("token", apiVUserCanLoginToken)
		params.Add("mobile", userId)
		params.Add("password", pwd)
		resp, err := http.PostForm(apiVUserCanLogin, params)
		lk.WarnOnErr("%v", err)
		if err != nil {
			cResult <- ResultExt{
				ok:  false,
				err: err,
			}
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		lk.WarnOnErr("%v", err)
		if err != nil {
			cResult <- ResultExt{
				ok:  false,
				err: err,
			}
			return
		}
		mData := make(map[string]any)
		lk.WarnOnErr("%v", json.Unmarshal(body, &mData))

		// fmt.Println(mData)

		if mData["code"].(float64) == 200 {
			if mData["status"] == "SUCCESS" {
				cResult <- ResultExt{
					ok:  true,
					err: nil,
				}
				return
			}
		}
		cResult <- ResultExt{
			ok:  false,
			err: fmt.Errorf("[%s] with password [%s] cannot login v-site", userId, pwd),
		}
	}()
	return cResult
}

func ExtUserLoginCheck(userId, pwd string) (bool, error) {
	select {
	case <-time.After(timeout * time.Second):
		return false, errors.New("timeout")
	case r := <-ExtUserLoginValidateAsync(userId, pwd):
		return r.ok, r.err
	}
}
