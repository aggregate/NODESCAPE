package main

import (
	"os"
	"nodescape/config"
)

func main() {
	in, _ := os.Open("test.conf")
	config.ReadConfig(in)
}
