package main

import (
	"fmt"
	"os"
	"os/exec"
)

// https://docs.invidious.io/api/#get-apiv1stats

func main() {
	if len(invidious) == 0 {
		fmt.Println("invidious instances do not existed")
		os.Exit(1)
	}
	instance = invidious[0]

	args := os.Args

	// UI to interact with playlist
	if len(args) == 1 {
		readVideosList()
		renderApp()
	}

	// add new channel to list
	if len(args) == 3 && args[1] == "add" {
		url := args[2]
		if url == "" {
			fmt.Println("empty channel url")
			os.Exit(1)
		}
		addChannel(url)
	}
}

func mpv(v Video) {
	app.Stop()
	fmt.Printf("playing %v", v.Title)
	cmd := exec.Command("mpv", "https://www.youtube.com/watch?v="+v.VideoID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
	renderApp()
}

func changeInstance() {
	if instanceTry > len(invidious)-1 {
		fmt.Println("tried all instance")
		os.Exit(1)
	}
	instance = invidious[instanceTry]
	fmt.Println("invidious instance:", instance)
}
