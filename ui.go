package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func renderApp() {
	app = tview.NewApplication()
	renderPlaylist()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlA:
			selected = 0
			sortVideosByDate(toggleDate)
			toggleDate = !toggleDate
			renderPlaylist()
		case tcell.KeyCtrlS:
			selected = 0
			sortVideosByMostView(toggleView)
			toggleView = !toggleView
			renderPlaylist()
		case tcell.KeyCtrlD:
			selected = 0
			sortVideosByLength(toggleLength)
			toggleLength = !toggleLength
			renderPlaylist()
		case tcell.KeyCtrlU:
			selected = 0
			scanVideos()
		case tcell.KeyCtrlE:
			exportM3U(0, playlistFile)
		}
		return event
	})
	err = app.Run()
	if err != nil {
		panic(err)
	}
}

func renderPlaylist() {
	list := tview.NewList()
	app.SetRoot(list, true)
	app.SetFocus(list)
	for i, v := range videos {
		d := time.Duration(v.LengthSeconds * 1000000000)
		list.AddItem(fmt.Sprintf("%v| %s", i, v.Title),
			fmt.Sprintf("       %v, %v, %s, %s", v.Author, v.ViewCountText, v.PublishedText, d.String()), rune(0), func() {
				selected = list.GetCurrentItem()
				mpv(v)
			})
	}
	if selected > 0 {
		list.SetCurrentItem(selected)
	}
}
