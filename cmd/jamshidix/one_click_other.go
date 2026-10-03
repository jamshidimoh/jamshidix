//go:build !windows

package main

import "errors"

var errUnsupported = errors.New("Jamshidix supports Windows only")

func runOneClick() error             { return errUnsupported }
func stopSingBox() error             { return errUnsupported }
func isSingBoxRunning() bool         { return false }
func readClipboard() (string, error) { return "", errUnsupported }
func setAutostart(bool) error        { return errUnsupported }
func attachParentConsole()           {}
