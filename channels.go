package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

func addChannel(url string) {
	client := resty.New()
	var resp *resty.Response
	resp, err = client.R().Get("https://" + invidious[instanceIndex] + "/api/v1/search?q=" + url)
	if err != nil {
		fmt.Println("rest client err:", err)
		changeInstance()
		addChannel(url)
	}
	if resp.StatusCode() != 200 {
		fmt.Println("  Status Code:", resp.StatusCode())
		fmt.Println("  Status     :", resp.Status())
		changeInstance()
		addChannel(url)
	} else {
		var result []SearchResult
		err = json.Unmarshal(resp.Body(), &result)
		if err != nil {
			fmt.Println("Unmarshal err:", err)
			os.Exit(1)
		}
		for _, r := range result {
			if r.Type == "channel" {
				cn := Channel{
					Name: r.Author,
					Url:  r.ChannelHandle,
					Id:   r.AuthorID,
				}
				readChannels()
				for _, c := range channels {
					if c.Id == cn.Id {
						fmt.Println("channel existed in the list")
						os.Exit(0)
					}
				}
				channels = append(channels, cn)
				saveChannels()
				fmt.Println("channel added to the list")
				break
			}
		}
	}
}

func deleteChannel(url string) {
	readChannels()
	var (
		ns    []Channel
		found bool
	)
	for _, c := range channels {
		if strings.EqualFold(c.Url, url) {
			found = true
		} else {
			ns = append(ns, c)
		}
	}
	channels = ns
	saveChannels()
	if found {
		fmt.Println("channel", url, "deleted")
	} else {
		fmt.Println("channel", url, "do no existed")
	}
}

func readChannels() {
	if _, err := os.Stat(channelsList); err != nil {
		saveChannels()
	}
	file, err := os.ReadFile(channelsList)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, &channels)
	if err != nil {
		panic(err)
	}
}

func saveChannels() {
	jdata, err := json.Marshal(channels)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(channelsList, jdata, 0755)
	if err != nil {
		panic(err)
	}
}

func getVideos(id string) (vs Videos) {
	client := resty.New()
	var resp *resty.Response
	resp, err = client.R().Get("https://" + invidious[instanceIndex] + "/api/v1/channels/" + id + "/videos")
	if err != nil {
		fmt.Println("rest client err:", err)
		os.Exit(1)
	}
	if resp.StatusCode() != 200 {
		fmt.Println("  Status Code:", resp.StatusCode())
		fmt.Println("  Status     :", resp.Status())
		changeInstance()
		getVideos(id)
	} else {
		err = json.Unmarshal(resp.Body(), &vs)
		if err != nil {
			fmt.Println("Unmarshal err:", err)
			os.Exit(1)
		}
	}
	return
}

func scanVideos() {
	readChannels()
	videos = []Video{}
	app.Stop()
	fmt.Println("scanning videos from channels")
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	for _, c := range channels {
		wg.Add(1)
		go func(c Channel) {
			defer wg.Done()
			fmt.Println(c.Name)
			vs := getVideos(c.Id)
			for _, v := range vs.Videos {
				if v.IsUpcoming || v.Premium {
					continue
				}
				mu.Lock()
				videos = append(videos, v)
				mu.Unlock()
			}
		}(c)
	}
	wg.Wait()
	saveVideosList()
	renderApp()
}

func saveVideosList() {
	jdata, err := json.Marshal(videos)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(videosJson, jdata, 0755)
	if err != nil {
		panic(err)
	}
}

func readVideosList() {
	if _, err := os.Stat(videosJson); err != nil {
		scanVideos()
		return
	}
	file, err := os.ReadFile(videosJson)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, &videos)
	if err != nil {
		panic(err)
	}
}

func exportM3U(index int, location string) {
	var strs []string
	strs = append(strs, "#EXTM3U")
	for i := index; i < len(videos); i++ {
		v := videos[i]
		d := time.Duration(v.LengthSeconds * 1000000000)
		since := time.Since(time.Unix(int64(v.Published), 0)).Round(time.Minute)
		strs = append(strs, fmt.Sprintf("#EXTINF: %v", v.Title))
		strs = append(strs, fmt.Sprintf("#EXTINF: %v, %v, since %v ago, %v", v.Author, v.ViewCountText, since, d.String()))
		strs = append(strs, "https://www.youtube.com/watch?v="+v.VideoID)
		strs = append(strs, "")
	}
	f, err := os.Create(location)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, l := range strs {
		_, err = f.WriteString(l + "\n")
		if err != nil {
			panic(err)
		}
	}
}

func sortVideosByLength() {
	sort.Slice(videos, func(i, j int) bool {
		return toggleLength == (videos[i].LengthSeconds < videos[j].LengthSeconds)
	})
	if toggleLength {
		sortby = "longest"
	} else {
		sortby = "shortest"
	}
	renderPlaylist()
	toggleLength = !toggleLength
}

func sortVideosByMostView() {
	sort.Slice(videos, func(i, j int) bool {
		return toggleView == (videos[i].ViewCount > videos[j].ViewCount)
	})
	if toggleView {
		sortby = "most view"
	} else {
		sortby = "less view"
	}
	renderPlaylist()
	toggleView = !toggleView
}

func sortVideosByDate() {
	sort.Slice(videos, func(i, j int) bool {
		return toggleDate == (videos[i].Published > videos[j].Published)
	})
	if toggleDate {
		sortby = "newest"
	} else {
		sortby = "oldest"
	}
	renderPlaylist()
	toggleDate = !toggleDate
}

func sortVideosByChannel() {
	sort.Slice(videos, func(i, j int) bool {
		return toggleChannel == (videos[i].Author < videos[j].Author)
	})
	if toggleChannel {
		sortby = "channel A-Z"
	} else {
		sortby = "channel Z-A"
	}
	renderPlaylist()
	toggleChannel = !toggleChannel
}
