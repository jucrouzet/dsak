// Package version is the version of dsak and utilities.
package version

import "fmt"

// V is the version of dsak.
// It is set at build time by Makefile.
const V string = "0.0.0"

// Build is the build tag of dsak.
// It is set at build time by Makefile.
const Build string = "dev"

func GetFullVersion() string {
	return fmt.Sprintf("%s-%s", V, Build)
}
