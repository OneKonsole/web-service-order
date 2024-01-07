package main

import "os"

var a App

func main() {
	a.Initialize("root", "root", "order")

	a.Run("8010")

	defer a.MQChannel.Close()
	defer a.MQConnection.Close()

	os.Exit(0)
}
