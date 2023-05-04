package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	key "github.com/digisan/gotk/crypto"
	fd "github.com/digisan/gotk/file-dir"
	lk "github.com/digisan/logkit"
	u "github.com/digisan/user-mgr/user"
	"github.com/golang-jwt/jwt/v4"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/postfinance/single"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/wismed-web/vhub-api/server/api"
	userapi "github.com/wismed-web/vhub-api/server/api/user"
	_ "github.com/wismed-web/vhub-api/server/docs" // once `swag init`, comment it out
)

var (
	fHttp2 = false //
	port   = 1323  // note: keep same as below @host // swagger ip for local test : 192.168.31.8
)

func init() {
	lk.Log("starting...main")
	lk.WarnDetail(false)
}

// @title WISMED V-HUB API
// @version 1.0
// @description This is WISMED V-HUB backend-api server. Updated@ 05-01-2023 16:49:24
// @termsOfService
// @contact.name API Support
// @contact.url
// @contact.email
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host api.v-hub.link
// @BasePath
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name authorization
func main() {

	var (
		http2FlagPtr = flag.Bool("h2", false, "http2 mode?")
	)
	flag.Parse()

	fHttp2 = *http2FlagPtr

	// only one instance
	const dir = "./tmp-locker"
	fd.MustCreateDir(dir)
	one, err := single.New("echo-service", single.WithLockPath(dir))
	lk.FailOnErr("%v", err)
	lk.FailOnErr("%v", one.Lock())
	defer func() {
		lk.FailOnErr("%v", one.Unlock())
		os.RemoveAll(dir)
		lk.Log("Server Exited Successfully")
	}()

	// start service
	done := make(chan string)
	echoHost(done)
	lk.Log(<-done)
}

func waitShutdown(e *echo.Echo) {
	go func() {
		// defer Close Database // after closing echo, close db

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		lk.Log("Got Ctrl+C")

		// other clean-up before closing echo
		{
		}

		// shutdown echo
		lk.FailOnErr("%v", e.Shutdown(ctx)) // close echo at e.Shutdown
	}()
}

func echoHost(done chan<- string) {
	go func() {
		defer func() { done <- "Echo Shutdown Successfully" }()

		e := echo.New()
		defer e.Close()

		// Middleware
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.Use(middleware.BodyLimit("2G"))
		// CORS
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowCredentials: true,
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		}))

		// waiting for shutdown
		waitShutdown(e)

		// host /swagger/index.html
		e.GET("/swagger/*", echoSwagger.WrapHandler)

		// host static file, such as agreement.pdf etc.
		e.File("/agreement", "static/test.pdf")

		// host static directory for user uploaded stuff.
		e.Static("/assets", "data/user-space/")

		// groups WITHOUT middleware
		{
			api.SystemHandler(e.Group("/api/system"))
			api.SignHandler(e.Group("/api/user/pub"))
			api.FileHandler(e.Group("/api/file/pub"))
		}

		// other groups WITH middleware,
		groups := []string{
			"/api/admin",
			"/api/user/auth",
			"/api/file/auth",
			"/api/submit",
			"/api/retrieve",
			"/api/manage",
			"/api/bookmark",
			"/api/reply",
			"/api/interact",
		}
		handlers := []func(*echo.Group){
			api.AdminHandler,
			api.UserAuthHandler,
			api.FileAuthHandler,
			api.SubmitHandler,
			api.RetrieveHandler,
			api.ManageHandler,
			api.BookmarkHandler,
			api.ReplyHandler,
			api.InteractHandler,
		}
		for i, group := range groups {
			r := e.Group(group)
			// r.Use(echojwt.JWT(StrToConstBytes(u.TokenKey()))) // HS256
			r.Use(echojwt.WithConfig(echojwt.Config{
				KeyFunc: getKey,
			}))
			r.Use(ValidateToken)
			handlers[i](r)
		}

		// running...
		portStr := fmt.Sprintf(":%d", port)
		var err error
		if fHttp2 {
			err = e.StartTLS(portStr, "./cert/public.pem", "./cert/private.pem")
		} else {
			err = e.Start(portStr)
		}
		lk.FailOnErrWhen(err != http.ErrServerClosed, "%v", err)
	}()
}

func ValidateToken(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, claims, err := u.TokenClaimsInHandler(c)
		if err != nil {
			return err
		}
		invoker := u.ClaimsToUser(claims)
		// if invoker.ValidateToken(token.Raw) { // HS256
		if ok, err := invoker.ValidateToken(token.Raw, userapi.PubKey); ok && err == nil { // RSA
			return next(c)
		}
		return c.JSON(http.StatusUnauthorized, map[string]any{
			"message": "invalid or expired jwt",
		})
	}
}

func getKey(token *jwt.Token) (interface{}, error) {
	// lk.Warn("%s\n", token.Raw)
	return key.ParseRsaPublicKeyFromPemStr(string(userapi.PubKey))
}
