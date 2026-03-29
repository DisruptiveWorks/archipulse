package main

import "embed"

//go:embed all:web
var staticFiles embed.FS
