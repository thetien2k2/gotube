package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/DexterLB/mpvipc"
)

type Video struct {
	Title     string  `json:"title"`
	Url       string  `json:"url"`
	Channel   string  `json:"channel"`
	Duration  float64 `json:"duration"`
	Timestamp int64   `json:"timestamp"`
	ViewCount int     `json:"view_count"`
}

func worker(jobs <-chan Channel, result chan<- Channel) {
	for j := range jobs {
		fmt.Println("updatding", j.Channel, j.Id)
		url := j.ChannelUrl
		if url == "" {
			url = "https://www.youtube.com/channel/" + j.Id
		}
		nc, err := getChannel(url)
		if err != nil {
			fmt.Println(err)
		}
		result <- nc
	}
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
	var newChannelList []Channel

	numJobs := len(channels)
	jobs := make(chan Channel, numJobs)
	result := make(chan Channel, numJobs)
	for range 5 {
		go worker(jobs, result)
	}
	for _, c := range channels {
		jobs <- c
	}
	close(jobs)
	for range numJobs {
		nc := <-result
		if len(nc.Videos) > 0 {
			for i, v := range nc.Videos {
				v.Channel = nc.Channel
				nc.Videos[i] = v
			}
			videos = append(videos, nc.Videos...)
			nc.Videos = []Video{}
			newChannelList = append(newChannelList, nc)
		}
	}
	saveVideosList()
  channels = newChannelList
	saveChannelsList()
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
		strs = append(strs, fmt.Sprintf("#EXTINF: %v", v.Title))
		strs = append(strs, v.Url)
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

// func search(query string) {
// 	rq := make(map[string]string)
// 	rq["q"] = query
// 	rq["type"] = "video"
// 	ep := "/api/v1/search?"
// 	var resp *resty.Response
// 	for _, i := range instances {
// 		resp, err = restGet(i, ep, rq)
// 		if err == nil {
// 			break
// 		}
// 	}
// 	if err != nil {
// 		return
// 	}
// 	var result []SearchResult
// 	err = json.Unmarshal(resp.Body(), &result)
// 	if err != nil {
// 		fmt.Println("Unmarshal err:", err)
// 		os.Exit(1)
// 	}
// 	videos = []Video{}
// 	for _, r := range result {
// 		if r.Type == "video" {
// 			v := Video{
// 				Title:         r.Title,
// 				VideoID:       r.VideoID,
// 				Author:        r.Author,
// 				ViewCount:     r.ViewCount,
// 				ViewCountText: r.ViewCountText,
// 				LengthSeconds: r.LengthSeconds,
// 				Published:     r.Published,
// 				PublishedText: r.PublishedText,
// 			}
// 			videos = append(videos, v)
// 		}
// 	}
// }

func sortVideosByLength() {
	sort.Slice(videos, func(i, j int) bool {
		return toggleLength == (videos[i].Duration < videos[j].Duration)
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
		return toggleDate == (videos[i].Timestamp > videos[j].Timestamp)
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
		if videos[i].Channel == videos[j].Channel {
			return videos[i].Timestamp > videos[j].Timestamp
		} else {
			return toggleChannel == (videos[i].Channel < videos[j].Channel)
		}
	})
	if toggleChannel {
		sortby = "channel A-Z"
	} else {
		sortby = "channel Z-A"
	}
	renderPlaylist()
}

func mpv(v Video) {
	app.Stop()
	done := make(chan string)
	if continuous {
		err := exportM3U(selected, tmpPlaylist)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	go func() {
		var (
			cmd  *exec.Cmd
			args []string
		)
		if continuous {
			args = append(args, fmt.Sprintf("--playlist=%s", tmpPlaylist))
		} else {
			args = append(args, v.Url)
		}
		if audioOnly {
			args = append(args, "--vid=no")
		}
		args = append(args, fmt.Sprintf("--input-ipc-server=%v", socket))
		cmd = exec.Command("mpv", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		cmd.Wait()
		done <- "done"
	}()

	time.Sleep(time.Second)
	conn := mpvipc.NewConnection(socket)
	err := conn.Open()
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	if err == nil {
		events, stopListening := conn.NewEventListener()
		go func() {
			conn.WaitUntilClosed()
			stopListening <- struct{}{}
		}()
		for event := range events {
			if event.Name == "file-loaded" {
				name, err := conn.Get("filename")
				if err == nil {
					mpvFileLoaded(name.(string))
				}
			}
		}
	}

	<-done
	fmt.Print("\033]0;gotube\007")
	renderApp()
}

func mpvFileLoaded(url string) {
	for i, v := range videos {
		if strings.Contains(v.Url, url) {
			fmt.Printf("\033]0;%s\007", v.Title)
			selected = i
			break
		}
	}
}
