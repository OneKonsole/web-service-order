package main

import (
	"os"
)

var a App
var appConf AppConf

func main() {
	// Dirty trick to pass conf globally
	appConf.Initialize()
	a.AppConf = &appConf

	// Init database, field validators, etc...
	a.Initialize()

	a.Run()

	os.Exit(0)
}
