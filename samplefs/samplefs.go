// Package samplefs implements a sample embedded template filesystem for tests and examples for apitpl.
package samplefs

// Generate resource.go by [parcello](github.com/phogolabs/parcello) from ../testdata, used only in example_*.go
//go:generate parcello -r -d ../testdata

import "github.com/phogolabs/parcello"

// FS returns embedded filesystem
func FS() parcello.FileSystemManager {
	return parcello.Manager
}
