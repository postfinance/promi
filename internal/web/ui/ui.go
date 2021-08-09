package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var static embed.FS

// ReactApp contains the prometheus react app.
func ReactApp() (http.FileSystem, error) {
	react, err := fs.Sub(static, "static/react")
	if err != nil {
		return nil, err
	}

	return http.FS(react), nil
}
