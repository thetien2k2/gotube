package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
		app := tview.NewApplication()
		list := tview.NewList()
		renderPlaylist(list)
		app.SetRoot(list, true)
		app.SetFocus(list)
		app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyCtrlA:
				sortVideosByDate(toggleDate)
				toggleDate = !toggleDate
				list.Clear()
				renderPlaylist(list)
			case tcell.KeyCtrlS:
				sortVideosByMostView(toggleView)
				toggleView = !toggleView
				list.Clear()
				renderPlaylist(list)
			case tcell.KeyCtrlD:
				sortVideosByLength(toggleLength)
				toggleLength = !toggleLength
				list.Clear()
				renderPlaylist(list)
			case tcell.KeyCtrlU:
				list.Clear()
				list.AddItem("scanning videos from channels", "", rune(0), nil)
				scanVideos()
				list.Clear()
				renderPlaylist(list)
			case tcell.KeyCtrlE:
				exportM3U()
			}
			return event
		})
		err = app.Run()
		if err != nil {
			panic(err)
		}
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

func mpv(vid string) {
	cmd := exec.Command("mpv", "https://www.youtube.com/watch?v="+vid)
	cmd.Run()
}

func changeInstance() {
	if instanceTry > len(invidious)-1 {
		fmt.Println("tried all instance")
		os.Exit(1)
	}
	// instance = invidious[rand.IntN(len(invidious))]
	instance = invidious[instanceTry]
	fmt.Println("invidious instance:", instance)
}
