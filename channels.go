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
	err = readChannels()
	if err != nil && err != errNoChannel {
		fmt.Println(err)
		os.Exit(1)
	}

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

func deleteChannel(id string) {
	err = readChannels()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var (
		ns    []Channel
		found bool
	)
	for _, c := range channels {
		if strings.EqualFold(c.Url, id) || strings.EqualFold(c.Id, id) {
			found = true
		} else {
			ns = append(ns, c)
		}
	}
	channels = ns
	saveChannels()
	if found {
		fmt.Println("channel with url/id", id, "deleted")
	} else {
		fmt.Println("channel with url/id", id, "do no existed")
	}
}

func readChannels() error {
	if _, err := os.Stat(dataDir + "/" + channelsList); err != nil {
		saveChannels()
	}
	file, err := os.ReadFile(dataDir + "/" + channelsList)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &channels)
	if err != nil {
		return err
	}
	if len(channels) == 0 {
		return errNoChannel
	}
	return nil
}

func listChannels() {
	err = readChannels()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Channels in the list:")
	fmt.Println("---------------------")
	for _, c := range channels {
		fmt.Printf("url: %v | id: %v | name: %v\n", c.Url, c.Id, c.Name)
	}
}

func saveChannels() {
	jdata, err := json.Marshal(channels)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(dataDir+"/"+channelsList, jdata, 0755)
	if err != nil {
		panic(err)
	}
}

func getVideos(id string) (vs Videos) {
	client := resty.New()
	var resp *resty.Response
	resp, err = client.R().Get("https://" + invidious[instanceIndex] + "/api/v1/channels/" + id + "/videos")
	if err != nil {
		fmt.Println(err)
		changeInstance()
		getVideos(id)
	}
	if resp.StatusCode() != 200 {
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
	err = readChannels()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if app != nil {
		app.Stop()
	}
	videos = []Video{}

	fmt.Println("updating video from channels")

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	for _, c := range channels {
		wg.Add(1)
		go func(c Channel) {
			defer wg.Done()
			vs := getVideos(c.Id)
			for _, v := range vs.Videos {
				if v.IsUpcoming || v.Premium {
					continue
				}
				mu.Lock()
				videos = append(videos, v)
				mu.Unlock()
			}
			fmt.Println(c.Name, "...done")
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
	err = os.WriteFile(dataDir+"/"+videosJson, jdata, 0755)
	if err != nil {
		panic(err)
	}
}

// https://docs.invidious.io/api/#get-apiv1stats
func readVideosList() {
	if _, err := os.Stat(dataDir + "/" + videosJson); err != nil {
		scanVideos()
		return
	}
	file, err := os.ReadFile(dataDir + "/" + videosJson)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, &videos)
	if err != nil {
		panic(err)
	}
}

func exportM3U(index int, location string) error {
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
		return err
	}
	defer f.Close()
	for _, l := range strs {
		_, err = f.WriteString(l + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func search(query, region string) {
	client := resty.New()
	var resp *resty.Response
	request := client.R()
	request.SetQueryParam("q", query)
	request.SetQueryParam("region", region)
	resp, err = request.Get("https://" + invidious[instanceIndex] + "/api/v1/search?")
	if err != nil {
		fmt.Println("rest client err:", err)
		changeInstance()
		search(query, region)
	}
	if resp.StatusCode() != 200 {
		fmt.Println("  Status Code:", resp.StatusCode())
		fmt.Println("  Status     :", resp.Status())
		changeInstance()
		search(query, region)
	} else {
		var result []SearchResult
		err = json.Unmarshal(resp.Body(), &result)
		if err != nil {
			fmt.Println("Unmarshal err:", err)
			os.Exit(1)
		}
		videos = []Video{}
		for _, r := range result {
			if r.Type == "video" {
				v := Video{
					Title:         r.Title,
					VideoID:       r.VideoID,
					Author:        r.Author,
					ViewCountText: r.ViewCountText,
					LengthSeconds: r.LengthSeconds,
				}
				videos = append(videos, v)
			}
		}
	}
}

func sortVideosByLength() {
	sort.Slice(videos, func(i, j int) bool {
		return toggleLength == (videos[i].LengthSeconds < videos[j].LengthSeconds)
	})
	if toggleLength {
		sortby = "shortest"
	} else {
		sortby = "longest"
	}
	renderPlaylist()
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
}

func sortVideosByChannel() {
	sort.Slice(videos, func(i, j int) bool {
		if videos[i].Author == videos[j].Author {
			return videos[i].Published > videos[j].Published
		} else {
			return toggleChannel == (videos[i].Author < videos[j].Author)
		}
	})
	if toggleChannel {
		sortby = "channel A-Z"
	} else {
		sortby = "channel Z-A"
	}
	renderPlaylist()
}
