// goforever - processes management
// Copyright (c) 2013 Garrett Woodworth (https://github.com/gwoo).

// sphere-director - Ninja processes management
// Copyright (c) 2014 Ninja Blocks Inc. (https://github.com/ninjablocks).

package main

import (
	//"fmt"
	"testing"
)

func Test_main(t *testing.T) {
	if daemon.Name != "director" {
		t.Error("Daemon name is not director")
	}
	daemon.Args = []string{"foo"}
	daemon.start(daemon.Name)
	if daemon.Args[0] != "foo" {
		t.Error("First arg not foo")
	}
	daemon.find()
	daemon.stop()
}
