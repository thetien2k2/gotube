package main

import (
	"fmt"
	"os"
)

func main() {
	prepareDataDir()

	args := os.Args

	switch {
	case len(args) == 1:
		err := readInstances()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		readVideosList()
		renderApp()
	case len(args) == 3 && args[1] == "addc":
		url := args[2]
		if url == "" {
			fmt.Println("empty channel url")
			os.Exit(1)
		}
		addChannel(url)
	case len(args) == 3 && args[1] == "rmc":
		url := args[2]
		if url == "" {
			fmt.Println("empty channel url")
			os.Exit(1)
		}
		deleteChannel(url)
	case len(args) == 2 && args[1] == "lsc":
		listChannels()
	case len(args) == 2 && args[1] == "lsi":
    updateInstances()
		listInstances()
	case len(args) == 3 && args[1] == "addi":
		url := args[2]
		if url == "" {
			fmt.Println("empty invidious instance url")
			os.Exit(1)
		}
		addInstance(url)
	case len(args) == 3 && args[1] == "rmi":
		url := args[2]
		if url == "" {
			fmt.Println("empty invidious instance url")
			os.Exit(1)
		}
		deleteInstance(url)

	default:
		fmt.Println("unknown command")
	}
}

func prepareDataDir() {
	dataDir, err = os.UserConfigDir()
	if err != nil {
		dataDir, err = os.UserHomeDir()
		if err != nil {
			fmt.Println("unable locate user's home directory or config directory")
			fmt.Println(err)
			os.Exit(1)
		}
	}
	dataDir += "/gotube"
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
