package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/go-resty/resty/v2"
)

// https://docs.invidious.io/api/#get-apiv1stats

func main() {
	args := os.Args
	fmt.Println(args)
	if len(invidious) == 0 {
		fmt.Println("invidious instances do not existed")
		os.Exit(1)
	}
	instance = invidious[0]

	// add new channel to list
	if len(args) == 3 && args[1] == "add" {
		url := args[2]
		if url == "" {
			fmt.Println("empty channel url")
			os.Exit(1)
		}
		addChannel(url)
	}

	// create a playlist by get and sort all videos from channels
	if len(args) >= 2 && args[1] == "create" {
		var sortby string
		if len(args) == 3 {
			sortby = args[2]
		}
		readChannels()
		var (
			list   []string
			videos []Video
		)
		for _, c := range channels {
			fmt.Println(c.Name)
			vs := getVideos(c.Id)
			videos = append(videos, vs.Videos...)
		}
		switch sortby {
		case "view":
			sort.Slice(videos, func(i, j int) bool {
				return videos[i].ViewCount > videos[j].ViewCount
			})
		case "short":
			sort.Slice(videos, func(i, j int) bool {
				return videos[i].LengthSeconds < videos[j].LengthSeconds
			})
		case "long":
			sort.Slice(videos, func(i, j int) bool {
				return videos[i].LengthSeconds > videos[j].LengthSeconds
			})
		default:
			sort.Slice(videos, func(i, j int) bool {
				return videos[i].Published > videos[j].Published
			})
		}
		for _, v := range videos {
			fmt.Println(v.Title)
			fmt.Printf("-- %v, %v, %v\n", v.ViewCountText, v.PublishedText, v.LengthSeconds)
			list = append(list, "https://www.youtube.com/watch?v="+v.VideoID)
		}
		f, err := os.Create("playlist.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		for _, l := range list {
			_, err = f.WriteString(l + "\n")
			if err != nil {
				log.Fatal(err)
			}
		}
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

func changeInstance() {
	if instanceTry > len(invidious)-1 {
		fmt.Println("tried all instance")
		os.Exit(1)
	}
	// instance = invidious[rand.IntN(len(invidious))]
	instance = invidious[instanceTry]
	fmt.Println("invidious instance:", instance)
}

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
		log.Fatal(err)
	}
	err = json.Unmarshal(file, &channels)
	if err != nil {
		log.Fatal(err)
	}
}

func saveChannels() {
	jdata, err := json.Marshal(channels)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(channelsList, jdata, 0755)
	if err != nil {
		log.Fatal(err)
	}
}
