//go:build tools

// Package main contains ... nothing.
// Just here to ensure that the build dependencies tools are here.
package main

import (
	_ "github.com/leodido/go-conventionalcommits"
	_ "github.com/leodido/go-conventionalcommits/parser"
)
