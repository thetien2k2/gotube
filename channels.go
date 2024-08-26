package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

func addChannel(url string) {
	err = readChannels()
	if err != nil && err != errNoChannel {
		fmt.Println(err)
		os.Exit(1)
	}

	ep := "/api/v1/search?q=" + url
  var resp *resty.Response
	for _, i := range instances {
		resp, err = restGet(i.Url, ep, make(map[string]string))
		if err == nil {
			break
		} else {
			fmt.Println(err)
		}
	}
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
					fmt.Println("channel existed")
					os.Exit(0)
				}
			}
			channels = append(channels, cn)
			saveChannels()
			fmt.Println("channel added")
			break
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
		fmt.Println("channel with url/id", id, "removed")
	} else {
		fmt.Println("channel with url/id", id, "do not existed")
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
	fmt.Println("Added channels:")
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
