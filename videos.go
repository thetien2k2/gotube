package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/DexterLB/mpvipc"
	"github.com/go-resty/resty/v2"
)

func getVideos(id string) (vs Videos, err error) {
	ep := fmt.Sprintf("/api/v1/channels/%v/videos", id)
	var resp *resty.Response
	for _, i := range instances {
		resp, err = restGet(i, ep, make(map[string]string))
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}
	err = json.Unmarshal(resp.Body(), &vs)
	if err != nil {
		return
	}
	return
}

func worker(jobs <-chan Channel, results chan<- Videos) {
	for j := range jobs {
		fmt.Println("updating", j.Name)
		vs, err := getVideos(j.Id)
		if err != nil {
			fmt.Println("update", j.Name, "failed", err)
		}
		results <- vs
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

	var mu sync.Mutex
	numJobs := len(channels)
	jobs := make(chan Channel, numJobs)
	results := make(chan Videos, numJobs)
	for range 10 {
		go worker(jobs, results)
	}
	for _, c := range channels {
		jobs <- c
	}
	close(jobs)
	for range numJobs {
		vs := <-results
		for _, v := range vs.Videos {
			if v.IsUpcoming || v.Premium {
				continue
			}
			mu.Lock()
			videos = append(videos, v)
			mu.Unlock()
		}
	}
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

func search(query string) {
	rq := make(map[string]string)
	rq["q"] = query
	rq["type"] = "video"
	ep := "/api/v1/search?"
	var resp *resty.Response
	for _, i := range instances {
		resp, err = restGet(i, ep, rq)
		if err == nil {
			break
		}
	}
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
				ViewCount:     r.ViewCount,
				ViewCountText: r.ViewCountText,
				LengthSeconds: r.LengthSeconds,
				Published:     r.Published,
				PublishedText: r.PublishedText,
			}
			videos = append(videos, v)
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
			args = append(args, fmt.Sprintf("https://www.youtube.com/watch?v=%v", v.VideoID))
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

func mpvFileLoaded(id string) {
	id = strings.Replace(id, "watch?v=", "", -1)
	for i, v := range videos {
		if v.VideoID == id {
			fmt.Printf("\033]0;%s\007", v.Title)
			selected = i
			break
		}
	}
}
