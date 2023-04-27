package file

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	fm "github.com/digisan/file-mgr"
	. "github.com/digisan/go-generics/v2"
	lk "github.com/digisan/logkit"
	u "github.com/digisan/user-mgr/user"
	"github.com/labstack/echo/v4"
	"github.com/wismed-web/vhub-api/server/api/user"
)

// *** after implementing, register with path in 'file.go' *** //

// @Title    path content
// @Summary  get content under specific path.
// @Description
// @Tags     File
// @Accept   json
// @Produce  json
// @Param    ym    query string true "year-month, e.g. 2022-05"
// @Param    gpath query string true "group path, e.g. group1/group2/group3"
// @Success  200   "OK - get content successfully"
// @Failure  500   "Fail - internal error"
// @Router   /api/file/auth/path-content [get]
// @Security ApiKeyAuth
func PathContent(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname = invoker.UName
		ym    = c.QueryParam("ym")
		gpath = c.QueryParam("gpath")
	)

	if _, ok, _ := u.LoadActiveUser(uname); !ok {
		return c.String(http.StatusForbidden, fmt.Sprintf("invalid invoker status@[%s], dormant?", uname))
	}

	// fetch user space for valid login
	us, ok := user.MapUserSpace.Load(uname)
	if !ok || us == nil {
		return c.String(http.StatusInternalServerError, "login error for [pathcontent] @"+uname)
	}

	content := us.(*fm.UserSpace).PathContent(filepath.Join(ym, gpath))
	return c.JSON(http.StatusOK, content)
}

// @Title    file item
// @Summary  get  file-items by given path or id.
// @Description
// @Tags     File
// @Accept   json
// @Produce  json
// @Param    id   query string true "file ID (md5)"
// @Success  200  "OK - get file items successfully"
// @Failure  400  "Fail - incorrect query param id"
// @Failure  404  "Fail - not found"
// @Failure  500  "Fail - internal error"
// @Router   /api/file/auth/file-items [get]
// @Security ApiKeyAuth
func FileItems(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname = invoker.UName
		id    = c.QueryParam("id")
	)

	// fetch user space for valid login
	us, ok := user.MapUserSpace.Load(uname)
	if !ok || us == nil {
		return c.String(http.StatusInternalServerError, "login error for [fileitem] @"+uname)
	}

	fis, err := us.(*fm.UserSpace).FileItems(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if len(fis) == 0 {
		return c.JSON(http.StatusNotFound, fis)
	}

	return c.JSON(http.StatusOK, fis)
}

// @Title    upload-formfile
// @Summary  upload file action via form file input.
// @Description
// @Tags     File
// @Accept   multipart/form-data
// @Produce  json
// @Param    note   formData string false "note for uploading file; if file is image or video, 'crop:x,y,w,h' for cropping"
// @Param    addym  formData bool   true  "add /yyyy-mm/ into storage path"
// @Param    group0 formData string false "1st category for uploading file"
// @Param    group1 formData string false "2nd category for uploading file"
// @Param    group2 formData string false "3rd category for uploading file"
// @Param    file   formData file   true  "file path for uploading"
// @Success  200 "OK - return storage path"
// @Failure  400 "Fail - file param is incorrect"
// @Failure  500 "Fail - internal error"
// @Router   /api/file/auth/upload-formfile [post]
// @Security ApiKeyAuth
func UploadFormFile(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname = invoker.UName
		// Read form fields
		note   = c.FormValue("note")
		addym  = c.FormValue("addym")
		group0 = c.FormValue("group0")
		group1 = c.FormValue("group1")
		group2 = c.FormValue("group2")
	)

	// fetch user space for valid login
	us, ok := user.MapUserSpace.Load(uname)
	if !ok || us == nil {
		return c.String(http.StatusInternalServerError, "login error for [upload] @"+uname)
	}

	// Read file
	file, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// fmt.Println("note ---> ", note)

	ymFlag, ok := AnyTryToType[bool](addym)
	if !ok {
		return c.String(http.StatusBadRequest, "[addym] is bool type")
	}

	path, err := us.(*fm.UserSpace).SaveFormFile(file, note, ymFlag, group0, group1, group2)
	if err != nil {
		lk.Warn("UploadFormFile / SaveFormFile ERR: %v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// * root    path   	"data/user-space/"
	// * storage path   	"data/user-space/cdutwhu/2022-05/g0/g1/g2/document/github key.1652858188.txt"
	// * this    return 	"2022-05/g0/g1/g2/document/github key.1652858188.txt"
	// * future  access url "[ip:port]/[uname]/2022-05/g0/g1/g2/document/github key.1652858188.txt" // refer to 'e.Static' in main.go

	parts := strings.Split(path, "/")
	path = fmt.Sprintf("assets/%s/", uname) + strings.Join(parts[3:], "/") // refer to 'e.Static' in main.go
	return c.JSON(http.StatusOK, path)
}

// @Title    upload-bodydata
// @Summary  upload file action via body content.
// @Description
// @Tags     File
// @Accept   application/octet-stream
// @Produce  json
// @Param    fname  query string true  "filename for uploading data from body"
// @Param    note   query string false "note for uploading file; if file is image or video, 'crop:x,y,w,h' for cropping"
// @Param    addym  query bool   true  "add /yyyy-mm/ into storage path"
// @Param    group0 query string false "1st category for uploading file"
// @Param    group1 query string false "2nd category for uploading file"
// @Param    group2 query string false "3rd category for uploading file"
// @Param    data   body  string true  "file data for uploading" Format(binary)
// @Success  200 "OK - return storage path"
// @Failure  400 "Fail - file param is incorrect"
// @Failure  500 "Fail - internal error"
// @Router   /api/file/auth/upload-bodydata [post]
// @Security ApiKeyAuth
func UploadBodyData(c echo.Context) error {

	invoker, err := u.Invoker(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var (
		uname   = invoker.UName
		fName   = c.QueryParam("fname")
		note    = c.QueryParam("note")
		addym   = c.QueryParam("addym")
		group0  = c.QueryParam("group0")
		group1  = c.QueryParam("group1")
		group2  = c.QueryParam("group2")
		dataRdr = c.Request().Body
	)

	// fetch user space for valid login
	us, ok := user.MapUserSpace.Load(uname)
	if !ok || us == nil {
		return c.String(http.StatusInternalServerError, "login error for [upload] @"+uname)
	}

	if len(fName) == 0 {
		return c.String(http.StatusBadRequest, "file name is empty")
	}
	if dataRdr == nil {
		return c.String(http.StatusBadRequest, "body data is empty")
	}

	ymFlag, ok := AnyTryToType[bool](addym)
	if !ok {
		return c.String(http.StatusBadRequest, "[addym] is bool type")
	}

	path, err := us.(*fm.UserSpace).SaveFile(dataRdr, fName, note, ymFlag, group0, group1, group2)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	parts := strings.Split(path, "/")
	path = strings.Join(parts[3:], "/")
	return c.JSON(http.StatusOK, path)
}
