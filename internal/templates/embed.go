package templates

import (
	_ "embed"
)

//go:embed object.go.tmpl
var ObjectTemplate []byte
