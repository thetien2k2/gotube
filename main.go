package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// https://docs.invidious.io/api/#get-apiv1stats

func main() {
	if len(invidious) == 0 {
		fmt.Println("invidious instances do not existed")
		os.Exit(1)
	}

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

	// add new channel to list
	if len(args) == 3 && args[1] == "delete" {
		url := args[2]
		if url == "" {
			fmt.Println("empty channel url")
			os.Exit(1)
		}
		deleteChannel(url)
	}

}

func mpv(v Video) {
	app.Stop()
	fmt.Println()
	fmt.Printf("🔊 %v\n", v.Title)
	cmd := exec.Command("mpv", "https://www.youtube.com/watch?v="+v.VideoID)
	output := new(bytes.Buffer)
	cmd.Stdout = output
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	cmd.Wait()
	reason, err := regexp.Compile(`Exiting...\s\(.*\)`)
	if err != nil {
		panic(err)
	}
	r := reason.FindString(output.String())
	r = strings.Replace(r, "Exiting... (", "", -1)
	r = strings.Replace(r, ")", "", -1)
	if r == "End of file" {
		selected++
		mpv(videos[selected])
	}
	renderApp()
}

func changeInstance() {
	if instanceChange > (len(invidious)-1)*instanceRetry {
		fmt.Println("tried all instances")
		os.Exit(1)
	}
	instanceChange++
	instanceIndex++
	if instanceIndex == len(invidious) {
		instanceIndex = 0
	}
	fmt.Println("invidious instance:", instanceIndex)
}
