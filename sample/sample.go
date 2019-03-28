// Package sample implements a sample code for tests and examples for tpl2x.
package sample

//go:generate parcello -r -d ../testdata

import "github.com/phogolabs/parcello"

func FS() parcello.FileSystemManager {
	return parcello.Manager
}
