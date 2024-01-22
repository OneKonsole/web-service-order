package main

import (
	"fmt"
	"os"
)

var a App
var appConf AppConf

func main() {
	// Dirty trick to pass conf globally
	appConf.Initialize()
	a.AppConf = &appConf

	fmt.Print("\nInitializing app IN MAIN GO...\n")
	// Init database, field validators, etc...
	a.Initialize()

	a.Run()

	defer a.DB.Close()

	os.Exit(0)
}
