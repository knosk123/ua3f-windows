package main

import (
	"embed"
	"io/fs"
)

//go:embed web/*
var webAssets embed.FS

func webStaticFS() fs.FS {
	sub, err := fs.Sub(webAssets, "web")
	if err != nil {
		panic(err)
	}
	return sub
}
