package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func renderApp() {
	if app != nil {
		app.Stop()
	}
	app = tview.NewApplication()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case rune('q'):
			app.Stop()
			os.Exit(0)
		case rune('w'):
			if app != nil {
				app.Stop()
			}
			scanVideos()
			resetSort()
			renderApp()
		case rune('1'):
			resetSort()
			renderPlaylist()
		case rune('2'):
			renderChannels()
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
	for i, v := range playlist {
		d := time.Duration(int(v.Duration) * 1000000000)
		viewcount := message.NewPrinter(language.English).Sprintf("%d", v.ViewCount)
		list.AddItem(fmt.Sprintf("%v| %s", i, v.Title),
			fmt.Sprintf("       %v | %v views | %s | %s", v.Channel, viewcount, d.String(), time.Unix(v.Timestamp, 0).Format(time.DateTime)), rune(0), func() {
				selected = list.GetCurrentItem()
				mpv(v)
			})
	}
	if selected > 0 {
		list.SetCurrentItem(selected)
	}
	var txtcontinuos, txtao string
	if continuous {
		txtcontinuos = "[continuous playing]"
	}
	if audioOnly {
		txtao = "[audio only]"
	}
	frame = tview.NewFrame(list).
		AddText(fmt.Sprintf("sort by: %v", sortby), true, tview.AlignLeft, tcell.ColorGray).
		AddText("gotube - playlist", true, tview.AlignCenter, tcell.ColorLightCyan).
		AddText(fmt.Sprintf("%v %v", txtcontinuos, txtao), true, tview.AlignRight, tcell.ColorGray).
		AddText("(q)quit (w)update | (z)continuous (x)audio only | (r)reset (c)clear", false, tview.AlignLeft, tcell.ColorGray).
		AddText("sort: (a)date (s)view (d)length (f)channel", false, tview.AlignLeft, tcell.ColorGray)
	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case rune('a'):
			selected = 0
			sortPlaylistByDate()
			toggleDate = !toggleDate
		case rune('s'):
			selected = 0
			sortPlaylistByView()
			toggleView = !toggleView
		case rune('d'):
			selected = 0
			sortPlaylistByLength()
			toggleLength = !toggleLength
		case rune('r'):
			playlist = videosDb
			renderPlaylist()
		case rune('c'):
			playlist = []Entry{}
			renderPlaylist()
		case rune('f'):
			selected = 0
			sortPlaylistByChannel()
			toggleChannel = !toggleChannel
		case rune('z'):
			continuous = !continuous
			selected = list.GetCurrentItem()
			renderPlaylist()
		case rune('x'):
			audioOnly = !audioOnly
			selected = list.GetCurrentItem()
			renderPlaylist()
		}
		return event
	})
	app.SetRoot(frame, true).SetFocus(frame)
}

func renderChannels() {
	list = tview.NewList()
	list.ShowSecondaryText(false)
	for _, c := range channels {
		list.AddItem(fmt.Sprintf("%s", c.Channel), "", rune(0), func() {
			playlist = append(playlist, videosByChannel(c)...)
			resetSort()
		})
	}
	frame = tview.NewFrame(list).
		AddText("gotube - channels", true, tview.AlignCenter, tcell.ColorLightCyan).
		AddText("(q)quit (w)update (d)delete channel (a)add channel", false, tview.AlignLeft, tcell.ColorGray)
	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case rune('d'):
			if app != nil {
				app.Stop()
			}
			name, _ := list.GetItemText(list.GetCurrentItem())
			deleteChannel(name)
			renderApp()
			renderChannels()
		case rune('a'):
			if app != nil {
				app.Stop()
			}
			addChannel()
			time.Sleep(time.Second)
			renderApp()
			renderChannels()
		}
		return event
	})
	app.SetRoot(frame, true).SetFocus(frame)
}
