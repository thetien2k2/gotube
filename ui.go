package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func renderApp() {
	app = tview.NewApplication()
	pages = tview.NewPages()
	renderPlaylist()
	pages.HidePage("modal")
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
			err = exportM3U(0, playlistFile)
			msg := ""
			if err != nil {
				msg = err.Error()
			} else {
				msg = "export completed"
			}
			modal := tview.NewModal()
			pages.AddPage("modal", modal, false, true)
			pages.ShowPage("modal")
			modal.SetText(msg)
			modal.AddButtons([]string{"OK"})
			modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "OK" {
					pages.HidePage("modal")
				}
			})
			app.SetFocus(modal)
		case tcell.KeyCtrlR:
			continuous = !continuous
			selected = list.GetCurrentItem()
			renderPlaylist()
		}
		return event
	})
	pages.HidePage("modal")
	app.SetRoot(pages, true).SetFocus(list)
	err = app.Run()
	if err != nil {
		panic(err)
	}
}

func renderPlaylist() {
	list = tview.NewList()
	list.SetBorder(true)
	if sortby == "" {
		sortby = "natural"
	}
	list.SetTitle(fmt.Sprintf(" gotubeplaylist, sort by: %v, continuous playing: %v ", sortby, continuous))
	pages.AddPage("list", list, true, true)
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
