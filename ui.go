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
			sortVideosByDate()
		case tcell.KeyCtrlS:
			selected = 0
			sortVideosByMostView()
		case tcell.KeyCtrlD:
			selected = 0
			sortVideosByLength()
		case tcell.KeyCtrlF:
			selected = 0
			sortVideosByChannel()
		case tcell.KeyCtrlU:
			selected = 0
			scanVideos()
		case tcell.KeyCtrlE:
			exportM3U(0, playlistFile)
		case tcell.KeyCtrlR:
			continuous = !continuous
			renderPlaylist()
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
	list.SetBorder(true)
	if sortby == "" {
		sortby = "natural"
	}
	list.SetTitle(fmt.Sprintf(" gotubeplaylist, sort by: %v, continuous playing: %v ", sortby, continuous))
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
