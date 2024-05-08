package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func renderApp() {
	if app != nil {
		app.Stop()
	}
	app = tview.NewApplication()
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
			sortby = ""
			scanVideos()
		case tcell.KeyCtrlE:
			app.Stop()
			err = exportM3U(0, playlistFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				fmt.Println("export completed!")
				time.Sleep(time.Second)
				renderApp()
			}
		case tcell.KeyCtrlR:
			continuous = !continuous
			selected = list.GetCurrentItem()
			renderPlaylist()
		}
		return event
	})
	renderPlaylist()
	app.SetRoot(frame, true).SetFocus(frame)
	err = app.Run()
	if err != nil {
		panic(err)
	}
}

func renderPlaylist() {
	list = tview.NewList()
	if sortby == "" {
		sortby = "natural"
	}
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
	frame = tview.NewFrame(list).
		AddText(fmt.Sprintf("sort by: %v", sortby), true, tview.AlignLeft, tcell.ColorGray).
		AddText("gotubeplaylist", true, tview.AlignCenter, tcell.ColorLightCyan).
		AddText(fmt.Sprintf("continuous playing: %v", continuous), true, tview.AlignRight, tcell.ColorGray).
		AddText("(u) scan new video | (e) export to m3u | (r) continuous playing | (c) quit", false, tview.AlignLeft, tcell.ColorGray).
		AddText("sort: (a) date, (s) view, (d) length, (f) channel", false, tview.AlignLeft, tcell.ColorGray)
	app.SetRoot(frame, true).SetFocus(frame)
}
