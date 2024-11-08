package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

const (
	channelsList = "channels.json"
	videosJson   = "videos.json"
	playlistFile = "playlist.m3u"
	socket       = "/tmp/mpvsocket"
	tmpPlaylist  = "/tmp/gotube.m3u"
	ytdlp        = "/usr/bin/yt-dlp"
)

var (
	dataDir       string
	channels      []Channel
	videos        []Entry
	err           error
	toggleDate    bool
	toggleView    bool
	toggleLength  bool
	toggleChannel bool
	app           *tview.Application
	list          *tview.List
	frame         *tview.Frame
	selected      int
	continuous    bool
	audioOnly     bool
	sortby        string
	errNoChannel  = fmt.Errorf("please add some channels, using addc command")
)

func main() {
	prepareDataDir()
	args := os.Args

	switch {
	case len(args) == 1:
		readVideosList()
		renderApp()
	case len(args) == 3 && args[1] == "add":
		name := args[2]
		if name == "" {
			fmt.Println("empty channel url")
			os.Exit(1)
		}
		addChannel(name)
	case len(args) == 3 && args[1] == "rm":
		url := args[2]
		if url == "" {
			fmt.Println("empty channel id")
			os.Exit(1)
		}
		deleteChannel(url)
	case len(args) == 2 && args[1] == "ls":
		listChannels()
	default:
		fmt.Println("unknown command")
	}
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

func getChannel(url string) (c Channel, err error) {
	if !strings.HasPrefix(url, "https://www.youtube.com") {
		err = fmt.Errorf("invalid url")
	}
	var (
		cmd  *exec.Cmd
		args []string
	)
	args = append(args, "--flat-playlist")
	args = append(args, "--no-warnings")
	args = append(args, "--extractor-args")
	args = append(args, "youtubetab:approximate_date")
	args = append(args, "-J")
	args = append(args, "-s")
	args = append(args, url)
	cmd = exec.Command(ytdlp, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	err = json.Unmarshal(out, &c)
	if err != nil {
		err = fmt.Errorf("Unmarshal err: %v", err)
		c = Channel{}
		return
	}
	if c.Channel == "" {
		err = fmt.Errorf("%v bad channel", c.ChannelId)
		c = Channel{}
		return
	}
	return
}
