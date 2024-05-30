package main

import (
	"fmt"
	"os"
	"os/exec"
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

func mpv(v Video) {
	app.Stop()
	done := make(chan string)
	fmt.Printf("\033]0;%s\007", v.Title)
	if continuous {
		err := exportM3U(selected, tmpPlaylist)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	go func() {
		fmt.Println()
		fmt.Println("🔊", v.Title)
		var (
			cmd  *exec.Cmd
			args []string
		)
		if continuous {
			args = append(args, fmt.Sprintf("--playlist=%s", tmpPlaylist))
		} else {
			args = append(args, fmt.Sprintf("https://www.youtube.com/watch?v=%v", v.VideoID))
		}
		if audioOnly {
			args = append(args, "--vid=no")
		}
		cmd = exec.Command("mpv", args...)
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
	<-done
	fmt.Print("\033]0;gotubeplaylist\007")
	renderApp()
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
