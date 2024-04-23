package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/go-resty/resty/v2"
)

func addChannel(url string) {
	client := resty.New()
	var resp *resty.Response
	resp, err = client.R().Get("https://" + instance + "/api/v1/search?q=" + url)
	if err != nil {
		fmt.Println("rest client err:", err)
		os.Exit(1)
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
	resp, err = client.R().Get("https://" + instance + "/api/v1/channels/" + id + "/videos")
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
	for _, c := range channels {
		vs := getVideos(c.Id)
		for _, v := range vs.Videos {
			if v.IsUpcoming || v.Premium {
				continue
			}
			videos = append(videos, v)
		}
	}
	saveVideosList()
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

func exportM3U() {
	var strs []string
	strs = append(strs, "#EXTM3U")
	for _, v := range videos {
		d := time.Duration(v.LengthSeconds * 1000000000)
		strs = append(strs, fmt.Sprintf("#EXTINF: %v", v.Title))
		strs = append(strs, fmt.Sprintf("#EXTINF: %v, %v, %v", v.ViewCountText, v.PublishedText, d.String()))
		strs = append(strs, "https://www.youtube.com/watch?v="+v.VideoID)
		strs = append(strs, "")
	}
	f, err := os.Create(playlistFile)
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

func sortVideosByLength(shortFirst bool) {
	sort.Slice(videos, func(i, j int) bool {
		return shortFirst == (videos[i].LengthSeconds > videos[j].LengthSeconds)
	})
}

func sortVideosByMostView(moreFirst bool) {
	sort.Slice(videos, func(i, j int) bool {
		return moreFirst == (videos[i].ViewCount > videos[j].ViewCount)
	})
}

func sortVideosByDate(newFirst bool) {
	sort.Slice(videos, func(i, j int) bool {
		return newFirst == (videos[i].Published > videos[j].Published)
	})
}
