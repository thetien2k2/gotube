package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Channel struct {
	Id         string  `json:"id"`
	Channel    string  `json:"channel"`
	ChannelUrl string  `json:"channel_url"`
	Follower   int64   `json:"channel_follower_count"`
	Videos     []Video `json:"entries"`
}

func addChannel(url string) {
	err = readChannels()
	if err != nil && err != errNoChannel {
		fmt.Println(err)
		os.Exit(1)
	}
	nc, err := getChannel(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, c := range channels {
		if c.Id == nc.Id {
			fmt.Println("channel existed")
			os.Exit(0)
		}
	}
	nc.Videos = []Video{}
	channels = append(channels, nc)
	saveChannelsList()
	fmt.Println("channel added")
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
		if strings.EqualFold(c.Id, id) {
			found = true
		} else {
			ns = append(ns, c)
		}
	}
	channels = ns
	saveChannelsList()
	if found {
		fmt.Println("channel with id", id, "removed")
	} else {
		fmt.Println("channel with id", id, "do not existed")
	}
}

func readChannels() error {
	if _, err := os.Stat(dataDir + "/" + channelsList); err != nil {
		saveChannelsList()
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
	fmt.Println("Added channels:")
	fmt.Println("---------------------")
	for _, c := range channels {
		fmt.Printf("id: %v | name: %v \n", c.Id, c.Channel)
	}
}

func saveChannelsList() {
	jdata, err := json.Marshal(channels)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(dataDir+"/"+channelsList, jdata, 0755)
	if err != nil {
		panic(err)
	}
}
