package main

import (
	"fmt"

	"github.com/rivo/tview"
)

const (
	channelsList = "channels.json"
	videosJson   = "videos.json"
	playlistFile = "playlist.m3u"
	socket       = "/tmp/mpvsocket"
)

var (
	invidious = []string{
		"invidious.fdn.fr",
		"vid.puffyan.us",
		"iv.ggtyler.dev",
		"inv.tux.pizza",
		"yewtu.be",
		"yt.artemislena.eu",
		"yt.cdaut.de",
		"inv.in.projectsegfau.lt",
		"invidious.perennialte.ch",
	}
	dataDir        string
	channels       []Channel
	videos         []Video
	err            error
	instanceIndex  int
	instanceChange int
	instanceRetry  = len(invidious) * 3
	toggleDate     = true
	toggleView     = true
	toggleLength   = true
	toggleChannel  = true
	app            *tview.Application
	list           *tview.List
	pages          *tview.Pages
	selected       int
	continuous     bool
	sortby         string
	errNoChannel   = fmt.Errorf("please add some channels first")
)

type Channel struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Id   string `json:"id"`
}

type SearchResult struct {
	Type             string `json:"type"`
	Author           string `json:"author"`
	AuthorID         string `json:"authorId"`
	AuthorURL        string `json:"authorUrl"`
	AuthorVerified   bool   `json:"authorVerified"`
	AuthorThumbnails []struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"authorThumbnails,omitempty"`
	AutoGenerated   bool   `json:"autoGenerated,omitempty"`
	SubCount        int    `json:"subCount,omitempty"`
	VideoCount      int    `json:"videoCount,omitempty"`
	ChannelHandle   string `json:"channelHandle,omitempty"`
	Description     string `json:"description"`
	DescriptionHTML string `json:"descriptionHtml"`
	Title           string `json:"title,omitempty"`
	VideoID         string `json:"videoId,omitempty"`
	VideoThumbnails []struct {
		Quality string `json:"quality"`
		URL     string `json:"url"`
		Width   int    `json:"width"`
		Height  int    `json:"height"`
	} `json:"videoThumbnails,omitempty"`
	ViewCount     int    `json:"viewCount,omitempty"`
	ViewCountText string `json:"viewCountText,omitempty"`
	Published     int    `json:"published,omitempty"`
	PublishedText string `json:"publishedText,omitempty"`
	LengthSeconds int    `json:"lengthSeconds,omitempty"`
	LiveNow       bool   `json:"liveNow,omitempty"`
	Premium       bool   `json:"premium,omitempty"`
	IsUpcoming    bool   `json:"isUpcoming,omitempty"`
}

type Videos struct {
	Videos []Video `json:"videos"`
}

type Video struct {
	Type            string `json:"type"`
	Title           string `json:"title"`
	VideoID         string `json:"videoId"`
	Author          string `json:"author"`
	AuthorID        string `json:"authorId"`
	AuthorURL       string `json:"authorUrl"`
	AuthorVerified  bool   `json:"authorVerified"`
	VideoThumbnails []struct {
		Quality string `json:"quality"`
		URL     string `json:"url"`
		Width   int    `json:"width"`
		Height  int    `json:"height"`
	} `json:"videoThumbnails"`
	Description     string `json:"description"`
	DescriptionHTML string `json:"descriptionHtml"`
	ViewCount       int    `json:"viewCount"`
	ViewCountText   string `json:"viewCountText"`
	Published       int    `json:"published"`
	PublishedText   string `json:"publishedText"`
	LengthSeconds   int    `json:"lengthSeconds"`
	LiveNow         bool   `json:"liveNow"`
	Premium         bool   `json:"premium"`
	IsUpcoming      bool   `json:"isUpcoming"`
}
