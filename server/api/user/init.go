package user

import (
	"context"
	"os"
	"strings"
	"time"

	cfg "github.com/digisan/go-config"
	key "github.com/digisan/gotk/crypto"
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

	prvKey []byte
	PubKey []byte

	// init-admin users
	admins []string

	// default avatar & avatar type
	defaultAvatar     []byte
	defaultAvatarType string
)

func init() {

	lk.Log("starting...user")

	ctx, Cancel = context.WithCancel(context.Background())

	// *** init key pair ***
	rsaPrvKey, rsaPubKey := key.GenerateRsaKeyPair()

	prvKeyStr, err := key.ExportRsaPrivateKeyAsPemStr(rsaPrvKey, "")
	lk.FailOnErr("error @ export private key: %v", err)
	prvKey = []byte(prvKeyStr)

	PubKeyStr, err := key.ExportRsaPublicKeyAsPemStr(rsaPubKey, "")
	lk.FailOnErr("error @ export public key: %v", err)
	PubKey = []byte(PubKeyStr)

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
	si.SetOfflineTimeout(7200 * time.Second) // heartbeats(offline) checker timeout
	monitorOfflineUser(ctx)

	// monitor token expired
	u.SetTokenValidPeriod(14400 * time.Second)
	monitorUserTokenExpired(ctx)

	// load initial admin users
	cfg.Init("init-admin", false, "./init-admin.json")
	cfg.Use("init-admin")
	cfg.Show()
	admins = cfg.ValArr[string]("admin")

	// load default avatar from resource
	defaultAvatar, err = os.ReadFile("./res/avatar.png")
	lk.FailOnErr("%v", err)
	defaultAvatarType = "image/png"
}

func monitorOfflineUser(ctx context.Context) {
	cOffline := make(chan string, 4096)
	si.MonitorOffline(ctx, cOffline, nil)
	go func() {
		for offline := range cOffline {
			if so.Logout(offline) == nil {
				if user, ok := UserCache.Load(offline); ok {
					lk.Log("deleting token: [%v]", offline)
					user.(*u.User).DeleteToken()
					UserCache.Delete(offline)
					MapUserSpace.Delete(offline)
				}
				lk.Log("offline: [%v]", offline)
			}
		}
	}()
}

func monitorUserTokenExpired(ctx context.Context) {
	cExpired := make(chan string, 4096)
	u.MonitorTokenExpired(ctx, cExpired, func(uname string) error {
		lk.Log("[%s]'s session is expired\n", uname)
		if so.Logout(uname) == nil {
			if user, ok := UserCache.Load(uname); ok {
				lk.Log("deleting token: [%v]", uname)
				user.(*u.User).DeleteToken()
				UserCache.Delete(uname)
				MapUserSpace.Delete(uname)
			}
		}
		return nil
	})
}
