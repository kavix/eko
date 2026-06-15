package main

import (
	"embed"
)

//go:embed all:ui/out
var assets embed.FS

func main() {
	run()
}
