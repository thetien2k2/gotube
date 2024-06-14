package main

import (
	"fmt"
	"os"
)

func main() {
	if len(invidious) == 0 {
		fmt.Println("invidious instances do not existed")
		os.Exit(1)
	}

	prepareDataDir()

	args := os.Args

	switch {
	case len(args) == 1:
		readVideosList()
		renderApp()
	case len(args) == 3 && args[1] == "add":
		url := args[2]
		if url == "" {
			fmt.Println("empty channel url")
			os.Exit(1)
		}
		addChannel(url)
	case len(args) == 3 && (args[1] == "delete" || args[1] == "rm"):
		url := args[2]
		if url == "" {
			fmt.Println("empty channel url")
			os.Exit(1)
		}
		deleteChannel(url)
	case len(args) == 2 && (args[1] == "list" || args[1] == "ls"):
		listChannels()
	default:
		fmt.Println("unknown command")
	}
}
