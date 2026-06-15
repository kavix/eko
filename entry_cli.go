//go:build no_gui

package main

import (
	"eko/cmd"
)

func run() {
	cmd.UIAssets = assets
	cmd.Execute()
}
