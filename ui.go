package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
		switch event.Rune() {
		case rune('a'):
			selected = 0
			sortVideosByDate()
			toggleDate = !toggleDate
		case rune('s'):
			selected = 0
			sortVideosByMostView()
			toggleView = !toggleView
		case rune('d'):
			selected = 0
			sortVideosByLength()
			toggleLength = !toggleLength
		case rune('f'):
			selected = 0
			sortVideosByChannel()
			toggleChannel = !toggleChannel
		case rune('q'):
			app.Stop()
			os.Exit(0)
		case rune('w'):
			selected = 0
			sortby = ""
			scanVideos()
		case rune('e'):
			app.Stop()
			fmt.Print("search: ")
			reader := bufio.NewReader(os.Stdin)
			query, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}
			query = strings.ReplaceAll(query, "\n", "", )
			if query != "" {
				search(query)
			}
			renderApp()
		case rune('o'):
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
		case rune('r'):
			continuous = !continuous
			selected = list.GetCurrentItem()
			renderPlaylist()
		case rune('t'):
			audioOnly = !audioOnly
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
	var txtcontinuos, txtao string
	if continuous {
		txtcontinuos = "[continuous playing]"
	}
	if audioOnly {
		txtao = "[audio only]"
	}
	frame = tview.NewFrame(list).
		AddText(fmt.Sprintf("sort by: %v", sortby), true, tview.AlignLeft, tcell.ColorGray).
		AddText("gotube", true, tview.AlignCenter, tcell.ColorLightCyan).
		AddText(fmt.Sprintf("%v %v", txtcontinuos, txtao), true, tview.AlignRight, tcell.ColorGray).
		AddText("(q)quit (w)update (e)search (r)continuous (t)audio only (o)export", false, tview.AlignLeft, tcell.ColorGray).
		AddText("sort: (a)date (s)view (d)length (f)channel", false, tview.AlignLeft, tcell.ColorGray)
	app.SetRoot(frame, true).SetFocus(frame)
}
