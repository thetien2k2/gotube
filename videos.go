package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/DexterLB/mpvipc"
)

type Entry struct {
	Type      string  `json:"_type"`
	Title     string  `json:"title"`
	Url       string  `json:"url"`
	Channel   string  `json:"channel"`
	Duration  float64 `json:"duration"`
	Timestamp int64   `json:"timestamp"`
	ViewCount int     `json:"view_count"`
	Entries   []Entry `json:"entries"`
}

func worker(jobs <-chan Channel, result chan<- Channel) {
	for j := range jobs {
		fmt.Println(j.Channel)
		url := j.ChannelUrl
		if url == "" {
			url = "https://www.youtube.com/channel/" + j.ChannelId
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
	videosDb = []Entry{}
	var newChannelList []Channel
	numJobs := len(channels)
	jobs := make(chan Channel, numJobs)
	result := make(chan Channel, numJobs)
	numRoutine := runtime.NumCPU()
	if numRoutine <= 4 {
		numRoutine = 8
	}
	for range numRoutine {
		go worker(jobs, result)
	}
	for _, c := range channels {
		jobs <- c
	}
	close(jobs)
	for range numJobs {
		nc := <-result
		addEntries(nc.Entries, nc.Channel)
		nc.Entries = []Entry{}
		newChannelList = append(newChannelList, nc)
	}
	saveVideosDb()
	channels = newChannelList
	saveChannelsList()
}

func addEntries(es []Entry, channel string) {
	for _, e := range es {
		if e.Type == "playlist" {
			addEntries(e.Entries, channel)
		}
		if e.Type == "url" && e.Duration > 0 {
			e.Channel = channel
			videosDb = append(videosDb, e)
		}
	}
}

func saveVideosDb() {
	jdata, err := json.Marshal(videosDb)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(dataDir+"/"+videosJson, jdata, 0755)
	if err != nil {
		panic(err)
	}
}

func readVideosDb() {
	if _, err := os.Stat(dataDir + "/" + videosJson); err != nil {
		scanVideos()
		return
	}
	file, err := os.ReadFile(dataDir + "/" + videosJson)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, &videosDb)
	if err != nil {
		panic(err)
	}
}

func exportM3U(index int, location string) error {
	var strs []string
	strs = append(strs, "#EXTM3U")
	for i := index; i < len(videosDb); i++ {
		v := videosDb[i]
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

func sortPlaylistByLength() {
	sort.Slice(playlist, func(i, j int) bool {
		return toggleLength == (playlist[i].Duration < playlist[j].Duration)
	})
	if toggleLength {
		sortby = "shortest"
	} else {
		sortby = "longest"
	}
	renderPlaylist()
}

func sortPlaylistByView() {
	sort.Slice(playlist, func(i, j int) bool {
		return toggleView == (playlist[i].ViewCount > playlist[j].ViewCount)
	})
	if toggleView {
		sortby = "most view"
	} else {
		sortby = "less view"
	}
	renderPlaylist()
}

func sortPlaylistByDate() {
	sort.Slice(playlist, func(i, j int) bool {
		return toggleDate == (playlist[i].Timestamp > playlist[j].Timestamp)
	})
	if toggleDate {
		sortby = "newest"
	} else {
		sortby = "oldest"
	}
	renderPlaylist()
}

func sortPlaylistByChannel() {
	sort.Slice(playlist, func(i, j int) bool {
		if playlist[i].Channel == playlist[j].Channel {
			return playlist[i].Timestamp > playlist[j].Timestamp
		} else {
			return toggleChannel == (playlist[i].Channel < playlist[j].Channel)
		}
	})
	if toggleChannel {
		sortby = "channel A-Z"
	} else {
		sortby = "channel Z-A"
	}
	renderPlaylist()
}

func resetSort() {
	selected = 0
	sortby = ""
	toggleChannel = false
	toggleDate = false
	toggleView = false
	toggleLength = false
}

func mpv(v Entry) {
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
	for i, v := range videosDb {
		if strings.Contains(v.Url, url) {
			fmt.Printf("\033]0;%s\007", v.Title)
			selected = i
			break
		}
	}
}

func videosByChannel(c Channel) (es []Entry) {
	for i := range videosDb {
		if videosDb[i].Channel == c.Channel {
			es = append(es, videosDb[i])
		}
	}
	return
}
