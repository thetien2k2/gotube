package main

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

func renderPlaylist(list *tview.List) {
	for i, v := range videos {
		d := time.Duration(v.LengthSeconds * 1000000000)
		list.AddItem(fmt.Sprintf("%v| %s", i, v.Title),
			fmt.Sprintf("       %v, %v, %s", v.ViewCountText, v.PublishedText, d.String()), rune(0), func() {
				mpv(v.VideoID)
			})
	}
}
