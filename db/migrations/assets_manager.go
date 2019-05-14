package migrations

import (
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4/source"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	pkgerr "github.com/pkg/errors"
)

// A FilesList is a files filter function
type FilesList func(fs http.FileSystem, dir string) ([]string, error)

// GetFilesList returns a files list in a directory
func GetFilesList(fs http.FileSystem, dir string) ([]string, error) {

	d, err := fs.Open(dir)
	if err != nil {
		return nil, pkgerr.Wrapf(err, "failed to open directory '%s'", dir)
	}

	list, err := d.Readdir(-1)
	if err != nil {
		return nil, pkgerr.Wrapf(err, "failed to read directory files list '%s'", dir)
	}

	retval := make([]string, len(list))
	for i := range list {
		retval[i] = list[i].Name()
	}

	return retval, nil
}

//NewFileReaderHandler returns a handler function of a file reader
func NewFileReaderHandler(fs http.FileSystem, dir string) func(name string) ([]byte, error) {
	return func(name string) ([]byte, error) {

		if dir != "" {
			name = filepath.Join(dir, name)
		}

		f, err := fs.Open(name)
		if err != nil {
			return nil, pkgerr.Wrap(err, "open file")
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, pkgerr.Wrap(err, "read file")
		}

		return b, nil
	}
}

// NewAssetsDriver create a migrations driver with assets list
func NewAssetsDriver(fs http.FileSystem, dir string, getAssets FilesList) (source.Driver, error) {

	assetsList, err := getAssets(fs, dir)
	if err != nil {
		return nil, pkgerr.Wrap(err, "files for migrations assets driver")
	}

	src := bindata.Resource(assetsList, NewFileReaderHandler(fs, dir))

	driver, err := bindata.WithInstance(src)
	if err != nil {
		return nil, pkgerr.Wrap(err, "create migrations assets driver")
	}

	return driver, nil
}
