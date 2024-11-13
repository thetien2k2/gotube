package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Channel struct {
	Channel    string  `json:"channel"`
	ChannelId  string  `json:"channel_id"`
	ChannelUrl string  `json:"channel_url"`
	Follower   int64   `json:"channel_follower_count"`
	Entries    []Entry `json:"entries"`
}

func addChannel() {
	err := readChannels()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("please enter channel url: ")
	var url string
	_, err = fmt.Scan(&url)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	url = strings.ReplaceAll(url, "\n", "")
	nc, err := getChannel(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	for _, c := range channels {
		if c.ChannelId == nc.ChannelId {
			fmt.Println("channel existed")
			os.Exit(0)
		}
	}
	addEntries(nc.Entries, nc.Channel)
	nc.Entries = []Entry{}
	channels = append(channels, nc)
	saveVideosDb()
	saveChannelsList()
	fmt.Println("added channel", nc.Channel)
}

func deleteChannel(name string) {
	fmt.Printf("remove channel %s ?[y/n]: ", name)
	var answer string
	_, err := fmt.Scan(&answer)
	if err != nil {
		fmt.Println("unable read answer")
		return
	}
	if answer != "y" {
		return
	}
	ns := []Channel{}
	for _, c := range channels {
		if c.Channel != name {
			ns = append(ns, c)
		}
	}
	channels = ns
	saveChannelsList()
	fmt.Println("deleted channel", name)
}

func readChannels() error {
	if _, err := os.Stat(dataDir + "/" + channelsJson); err != nil {
		saveChannelsList()
	}
	file, err := os.ReadFile(dataDir + "/" + channelsJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &channels)
	if err != nil {
		return err
	}
	return nil
}

func listChannels() {
	err := readChannels()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Added channels:")
	fmt.Println("---------------------")
	for _, c := range channels {
		fmt.Printf("id: %v | name: %v \n", c.ChannelId, c.Channel)
	}
}

func saveChannelsList() {
	jdata, err := json.Marshal(channels)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(dataDir+"/"+channelsJson, jdata, 0755)
	if err != nil {
		panic(err)
	}
}

func getChannel(url string) (c Channel, err error) {
	if !strings.HasPrefix(url, "https://www.youtube.com") {
		err = fmt.Errorf("invalid url")
		return
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
		fmt.Println(string(out))
		return c, err
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
