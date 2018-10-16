package scaffold

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// IoWriter uses the default go ioutils package to write the files.
type IoWriter struct {
}

func (i *IoWriter) Write(path string, file File) error {
	dir := filepath.Dir(path)

	var dirPerms os.FileMode

	if file.DirPermissions != 0 {
		dirPerms = file.DirPermissions
	} else {
		dirPerms = 0755
	}

	os.MkdirAll(dir, dirPerms)

	err := ioutil.WriteFile(path, []byte(file.Template), file.Permissions)

	return err
}
