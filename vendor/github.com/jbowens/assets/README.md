# assets

The _assets_ package provides a common interface for accessing and serving web
assets in Go applications.

Assets are exposed to the application as an `assets.Bundle`. Bundles can be backed
by the local file system, [data embedded in the binary](https://github.com/jteeuwen/go-bindata), etc.

[![GoDoc](https://godoc.org/github.com/jbowens/assets?status.svg)](https://godoc.org/github.com/jbowens/assets)
