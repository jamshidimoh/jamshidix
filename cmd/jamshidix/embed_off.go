//go:build !embedsb

package main

// Without the embedsb build tag the archive is downloaded on first run.
var embeddedSingBoxZip []byte
