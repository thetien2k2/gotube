package main

import (
	"fmt"
	"os"

	"github.com/rivo/tview"
)

const (
	channelsJson = "channels.json"
	videosJson   = "videos.json"
	playlistFile = "playlist.m3u"
	socket       = "/tmp/mpvsocket"
	tmpPlaylist  = "/tmp/gotube.m3u"
	ytdlp        = "/usr/bin/yt-dlp"
)

var (
	dataDir       string
	channels      []Channel
	videosDb      []Entry
	playlist      []Entry
	err           error
	toggleDate    bool
	toggleView    bool
	toggleLength  bool
	toggleChannel bool
	continuous    bool
	audioOnly     bool
	sortby        string
	app           *tview.Application
	list          *tview.List
	frame         *tview.Frame
	selected      int
)

func main() {
	prepareDataDir()
	readChannels()
	readVideosDb()
	renderApp()
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
	dataDir += "/gotube"
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
