//go:build !windows

package main

import "errors"

var errOneClickRelaunch = errors.New("elevated instance launched")

func runOneClick() error { return errors.New("Jamshidix supports Windows only") }
func stopSingBox() error { return errors.New("Jamshidix supports Windows only") }
