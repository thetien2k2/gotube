package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	app         *tview.Application
	frame       *tview.Frame
	currentView string
)

func initApp() {
	app = tview.NewApplication()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case rune('q'):
			app.Stop()
			os.Exit(0)
		case rune('w'):
			scanVideos()
			resetSort()
		case rune('1'):
			currentView = "playlist"
			renderPlaylist()
		case rune('2'):
			currentView = "channels"
			renderChannels()
		}
		return event
	})
	if currentView == "channels" {
		renderChannels()
	} else {
		renderPlaylist()
	}
	err := app.Run()
	if err != nil {
		panic(err)
	}
}

var selected int

func renderPlaylist() {
	list := tview.NewList()
	if sortby == "" {
		sortby = "no sort"
	}
	for i, v := range playlist {
		d := time.Duration(int(v.Duration) * 1000000000)
		viewcount := message.NewPrinter(language.English).Sprintf("%d", v.ViewCount)
		list.AddItem(fmt.Sprintf("%v| %s", i, v.Title),
			fmt.Sprintf("       %v | %v views | %s | %s", v.Channel, viewcount, d.String(), time.Unix(v.Timestamp, 0).Format(time.DateTime)), rune(0), func() {
				selected = i
				var es []Entry
				if continuous {
					for i := selected; i < len(playlist); i++ {
						es = append(es, playlist[i])
					}
				} else {
					es = append(es, playlist[selected])
				}
				mpv(es)
				fmt.Println(selected, list.GetItemCount())
				os.Exit(0)
			})
	}
	list.SetCurrentItem(selected)
	var txtcontinuos, txtao string
	if continuous {
		txtcontinuos = " continuous"
	}
	if audioOnly {
		txtao = " audio"
	}
	frame = tview.NewFrame(list).
		AddText("gotube playlist", true, tview.AlignLeft, tcell.ColorLightCyan).
		AddText(fmt.Sprintf("%v%v%v", sortby, txtcontinuos, txtao), true, tview.AlignRight, tcell.ColorGray).
		AddText("(q)quit (w)update (1)playlist (2)channels", false, tview.AlignLeft, tcell.ColorGray).
		AddText("(z)continuous (x)audio only | (r)reset (c)clear", false, tview.AlignLeft, tcell.ColorGray).
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
		case rune('f'):
			selected = 0
			sortPlaylistByChannel()
			toggleChannel = !toggleChannel
		case rune('r'):
			playlist = videosDb
			renderPlaylist()
		case rune('c'):
			playlist = []Entry{}
			renderPlaylist()
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
	sort.Slice(channels, func(i, j int) bool {
		return strings.Compare(channels[i].Channel, channels[j].Channel) < 0
	})
	list := tview.NewList()
	list.ShowSecondaryText(false)
	for _, c := range channels {
		list.AddItem(fmt.Sprintf("%s", c.Channel), "", rune(0), func() {
			playlist = append(playlist, videosByChannel(c)...)
			resetSort()
		})
	}
	frame = tview.NewFrame(list).
		AddText("gotube channels", true, tview.AlignLeft, tcell.ColorLightCyan).
		AddText("(q)quit (w)update (1)playlist (2)channels", false, tview.AlignLeft, tcell.ColorGray).
		AddText("(a)add channel (d)delete channel", false, tview.AlignLeft, tcell.ColorGray)
	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case rune('d'):
			if app != nil {
				app.Stop()
			}
			name, _ := list.GetItemText(list.GetCurrentItem())
			deleteChannel(name)
			time.Sleep(time.Second)
			initApp()
		case rune('a'):
			if app != nil {
				app.Stop()
			}
			addChannel()
			time.Sleep(time.Second)
			initApp()
		}
		return event
	})
	app.SetRoot(frame, true).SetFocus(frame)
}
