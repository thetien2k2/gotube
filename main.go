package main

import (
	"fmt"
	"os"
)

const (
	channelsJson = "channels.json"
	videosJson   = "videos.json"
	playlistFile = "playlist.m3u"
	socket       = "/tmp/mpvsocket"
	tmpPlaylist  = "/tmp/gotube.m3u"
	ytdlp        = "/usr/bin/yt-dlp"
)

var dataDir string

func main() {
	prepareDataDir()
	readChannels()
	readVideosDb()
	initApp()
}

func prepareDataDir() {
	var err error
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
