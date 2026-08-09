package templates

import "embed"

//go:embed index.html fragments/*
var Files embed.FS
