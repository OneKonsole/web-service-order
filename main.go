package main

import "os"

var a App

func main() {
	a.Initialize("root", "root", "order")

	a.Run("8010")

	os.Exit(0)
}
