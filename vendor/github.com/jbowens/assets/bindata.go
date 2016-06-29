package assets

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// Bindata returns an assets bundle for the assets embedded in the
// binary via the `go-bindata` command.
func Bindata(bindata BindataPkg) (Bundle, error) {
	return bin{
		Handler: http.FileServer(BindataFileSystem(bindata)),
		paths:   bindata.AssetNames(),
	}, nil
}

type bin struct {
	http.Handler
	paths []string
}

func (b bin) RelativePaths() []string { return b.paths }

// BindataFileSystem constructs a http.FileSystem implementation that
// accesses assets stored within the application binary through go-bindata:
// https://github.com/jteeuwen/go-bindata
//
// It differs from the popular `go-bindata-assetfs` package in that it
// performs no code generation, allowing you to always use the default
// `go-bindata` command.
func BindataFileSystem(bindata BindataPkg) http.FileSystem {
	return &binFS{
		bindata: bindata,
	}
}

// BindataPkg is a wrapper around the functions genereated by the go-bindata
// command. Construct a BindataPkg by passing in your local package's
// code-generated functions:
//
//    bindataPkg := assets.BindataPkg{
//            Asset:      Asset,
//            AssetDir:   AssetDir,
//            AssetInfo:  AssetInfo,
//            AssetNames: AssetNames,
//    }
//
type BindataPkg struct {
	Asset      func(string) ([]byte, error)
	AssetDir   func(name string) ([]string, error)
	AssetInfo  func(string) (os.FileInfo, error)
	AssetNames func() []string
}

type binFS struct {
	bindata BindataPkg
}

// Open implements http.FileSystem and accesses the binary data embedded
// within the application binary.
func (fs *binFS) Open(name string) (http.File, error) {
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}

	info, err := fs.bindata.AssetInfo(name)
	if err != nil {
		return nil, err
	}

	f := bindataFile{
		fs:   fs,
		path: name,
		info: info,
	}

	b, err := fs.bindata.Asset(name)
	f.Reader = bytes.NewReader(b)
	// If successful, treat `name` as a normal file.
	if err == nil {
		return f, nil
	}

	// If it's not a normal file, it might still be a directory.
	children, err := fs.bindata.AssetDir(name)
	if err != nil {
		return nil, err
	}
	dir := bindataDir{
		bindataFile: f,
		dirents:     make([]os.FileInfo, 0, len(children)),
	}
	for _, c := range children {
		info, err := fs.bindata.AssetInfo(filepath.Join(name, c))
		if err != nil {
			return nil, err
		}
		dir.dirents = append(dir.dirents, info)
	}
	return dir, err
}

type bindataFile struct {
	*bytes.Reader
	fs   *binFS
	path string
	info os.FileInfo
}

func (f bindataFile) Close() error { return nil }

func (f bindataFile) Stat() (os.FileInfo, error) { return f.info, nil }

func (f bindataFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("%s is not a directory", f.path)
}

type bindataDir struct {
	bindataFile
	dirents []os.FileInfo
}

func (f bindataDir) Readdir(count int) (ret []os.FileInfo, err error) {
	if count >= len(f.dirents) {
		ret = f.dirents
		f.dirents = nil
		return ret, nil
	}
	ret = f.dirents[:count]
	f.dirents = f.dirents[count:]
	return ret, nil
}
