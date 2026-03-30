package main

import "embed"

//go:embed all:ui/dist
var staticFiles embed.FS
