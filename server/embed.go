package server

import (
	"embed"
)

//go:embed html/*
var content embed.FS
