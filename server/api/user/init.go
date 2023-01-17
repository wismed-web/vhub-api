package user

import (
	"context"
	"strings"
	"time"

	cfg "github.com/digisan/go-config"
	lk "github.com/digisan/logkit"
	si "github.com/digisan/user-mgr/sign-in"
	so "github.com/digisan/user-mgr/sign-out"
	su "github.com/digisan/user-mgr/sign-up"
	u "github.com/digisan/user-mgr/user"
	vf "github.com/digisan/user-mgr/user/valfield"
)

var (
	ctx    context.Context
	Cancel context.CancelFunc

	// init-admin users
	admins []string
)

func init() {

	lk.Log("starting...user")

	// set user db dir, activate ***[UserDB]***
	u.InitDB("./data/db-user")

	// set user validator
	su.SetValidator(map[string]func(o, v any) u.ValRst{
		vf.AvatarType: func(o, v any) u.ValRst {
			ok := v == "" || strings.HasPrefix(v.(string), "image/")
			return u.NewValRst(ok, "avatarType must have prefix - 'image/'")
		},
	})

	// monitor active users
	ctx, Cancel = context.WithCancel(context.Background())
	monitorUser(ctx, 7200*time.Second) // heartbeats checker timeout

	// load initial admin users
	cfg.Init("init-admin", false, "./init-admin.json")
	cfg.Use("init-admin")
	cfg.Show()
	admins = cfg.ValArr[string]("admin")
}

func monitorUser(ctx context.Context, offlineTimeout time.Duration) {
	cInactive := make(chan string, 4096)
	si.MonitorInactive(ctx, cInactive, offlineTimeout, nil)
	go func() {
		for inactive := range cInactive {
			if so.Logout(inactive) == nil {
				if claims, ok := MapUserClaims.Load(inactive); ok {
					lk.Log("delete token: [%v]", inactive)
					claims.(*u.UserClaims).DeleteToken()
					MapUserClaims.Delete(inactive)
				}
				lk.Log("offline: [%v]", inactive)
			}
		}
	}()
}
