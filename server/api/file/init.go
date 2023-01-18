package file

import fm "github.com/digisan/file-mgr"

func init() {
	// set user file space & file item db space
	fm.InitFileMgr("./data/")
	// set doing self file storage check when saving
	fm.OptCheckOnSave(true)
}
