package assets

import "net/http"

// Bundle represents a set of assets. It provides access to the current
// relative paths that the assets are served from and can be used as a
// http.Handler that will serve the assets at those same paths.
type Bundle interface {
	http.Handler
	RelativePaths() []string
}

// Bundler provides a common interface for types that can create asset
// bundles.
type Bundler interface {
	Bundle(directory string) (Bundle, error)
}

type bundlerFunc func(directory string) (Bundle, error)

// Bundle implements the Bundler interface by invoking bf.
func (bf bundlerFunc) Bundle(directory string) (Bundle, error) {
	return bf(directory)
}
