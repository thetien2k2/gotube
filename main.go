package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/DexterLB/mpvipc"
)

func main() {
	if len(invidious) == 0 {
		fmt.Println("invidious instances do not existed")
		os.Exit(1)
	}

	prepareDataDir()
	args := os.Args

	// Main UI
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

	// delete channel from list
	if len(args) == 3 && args[1] == "delete" {
		url := args[2]
		if url == "" {
			fmt.Println("empty channel url")
			os.Exit(1)
		}
		deleteChannel(url)
	}

	// list channels
	if len(args) == 2 && args[1] == "list" {
		listChannels()
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
		if selected > len(videos)-1 {
			selected = 0
		}
		mpv(videos[selected])
	} else {
		renderApp()
	}
}

func changeInstance() error {
	if instanceChange > (len(invidious)-1)*instanceRetry {
		app.Stop()
		fmt.Println("tried all instances but no answer")
		os.Exit(1)
	}
	instanceChange++
	instanceIndex++
	if instanceIndex == len(invidious) {
		instanceIndex = 0
	}
	return nil
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
	dataDir += "/gotubeplaylist"
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
