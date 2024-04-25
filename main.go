package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/DexterLB/mpvipc"
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
	done := make(chan string)
	endReason := ""
	go func() {
		fmt.Println()
		fmt.Println("🔊", v.Title)
		cmd := exec.Command("mpv", fmt.Sprintf("https://www.youtube.com/watch?v=%v", v.VideoID), fmt.Sprintf("--input-ipc-server=%v", socket))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		cmd.Wait()
		done <- "done"
	}()
	time.Sleep(time.Second)
	conn := mpvipc.NewConnection(socket)
	err := conn.Open()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	if err == nil {
		events, stopListening := conn.NewEventListener()
		go func() {
			conn.WaitUntilClosed()
			stopListening <- struct{}{}
		}()
		for event := range events {
			if event.Name == "end-file" {
				endReason = event.Reason
			}
		}
	}
	<-done
	if endReason == "eof" && continuous {
		selected++
		mpv(videos[selected])
	} else {
		renderApp()
	}
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
