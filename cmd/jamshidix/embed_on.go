//go:build embedsb

package main

import _ "embed"

// Release builds embed the pinned official sing-box archive (see
// .github/workflows/build-windows.yml), making the EXE fully self-contained.
//
//go:embed assets/sing-box-windows-amd64.zip
var embeddedSingBoxZip []byte
