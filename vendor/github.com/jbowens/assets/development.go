package assets

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DevelopmentBundler can create asset bundles in development mode.
var DevelopmentBundler Bundler

func init() {
	DevelopmentBundler = bundlerFunc(Development)
}

// Development returns an assets bundle for the assets in the the
// provided directory. It will pull the latest copy of the file from the
// filesystem on every request and is intended to be used during development.
//
// The relative paths returned by a development bundle will be suffixed with
// a fingerprint of the file size and last modified date.
func Development(dir string) (Bundle, error) {
	dir = filepath.Clean(dir)
	return dev{
		root: dir,
		fs:   http.FileServer(http.Dir(dir)),
	}, nil
}

type dev struct {
	root string
	fs   http.Handler
}

// RelativePaths provides the relative paths of all the assets in the bundle.
// Typically, RelativePaths is called on page load for specifying the location
// of stylesheets and scripts.
func (d dev) RelativePaths() (relativePaths []string) {
	_ = filepath.Walk(d.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(d.root, path)
		if err != nil {
			return nil
		}
		ext := filepath.Ext(relative)
		fingerprinted := strings.TrimSuffix(relative, ext) + "-" + fingerprint(info) + ext
		relativePaths = append(relativePaths, fingerprinted)
		return nil
	})
	return relativePaths
}

// ServeHTTP implements http.Handler and serves the assets in the bundle.
func (d dev) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	stripped := stripFingerprint(req.URL.Path)
	// Serve the file with the fingerprint stripped if it exists.
	if stripped != "" {
		if _, err := os.Stat(filepath.Join(d.root, stripped)); err == nil {
			req.URL.Path = stripped
		}
	}

	// Prevent caching during development.
	rw.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	rw.Header().Set("Pragma", "no-cache")
	rw.Header().Set("Expires", "0")

	d.fs.ServeHTTP(rw, req)
}

func stripFingerprint(p string) string {
	// Try to pull out the fingerprint hash.
	withoutExt := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
	idx := strings.LastIndexAny(withoutExt, "-")
	if idx == -1 {
		return ""
	}
	if _, err := hex.DecodeString(withoutExt[idx+1:]); err != nil {
		return ""
	}
	withoutFingerprint := withoutExt[:idx] + filepath.Ext(p)
	return filepath.Join(filepath.Dir(p), withoutFingerprint)
}

// fingerprint generates a hash of a file's Last Modified date and the size of
// the file. This technique isn't as reliable as hashing the entire body of
// the file, but it's good enough.
func fingerprint(info os.FileInfo) string {
	hasher := md5.New()
	_, _ = io.WriteString(hasher, info.Name())
	_ = binary.Write(hasher, binary.LittleEndian, info.ModTime().UnixNano())
	h := hasher.Sum(nil)
	return hex.EncodeToString(h[:4])
}
