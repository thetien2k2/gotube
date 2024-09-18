package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/go-resty/resty/v2"
)

func updateInstances() {
	err = readInstances()
	if err != nil && err != errNoInstance {
		fmt.Println(err)
		os.Exit(1)
	}
	var nis []Instance
	for _, i := range instances {
		var resp *resty.Response
		resp, err = restGet(i.Url, "/api/v1/stats", make(map[string]string))
		if err != nil {
			i.Online = false
      continue
		} 
    i.Online = true
    i.Ping = resp.Time().Seconds()
    nis = append(nis, i)
	}

	sort.Slice(nis, func(i, j int) bool {
		if nis[i].Online && !nis[j].Online {
			return true
		}
		return nis[i].Ping < nis[j].Ping
	})

	instances = nis
	saveInstances()
}

func addInstance(url string) {
	err = readInstances()
	if err != nil && err != errNoInstance {
		fmt.Println(err)
		os.Exit(1)
	}
	var resp *resty.Response
	resp, err = restGet(url, "/api/v1/stats", make(map[string]string))
	if err != nil {
		fmt.Println("rest client err:", err)
		os.Exit(1)
	}
	isInviduos := strings.ContainsAny(string(resp.Body()), "invidious")
	if !isInviduos {
		fmt.Println("this is not a Invidious instance")
		os.Exit(1)
	}
	for _, i := range instances {
		if i.Url == url {
			fmt.Println("instance existed")
			os.Exit(1)
		}
	}
	instances = append(instances, Instance{Url: url})
	saveInstances()
	fmt.Println("instance added")
}

func deleteInstance(url string) {
	err = readInstances()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var (
		ns    []Instance
		found bool
	)
	for _, c := range instances {
		if strings.EqualFold(c.Url, url) {
			found = true
		} else {
			ns = append(ns, c)
		}
	}
	if found {
		instances = ns
		saveInstances()
		fmt.Println(url, "removed")
	} else {
		fmt.Println(url, "do not existed")
	}
}

func readInstances() error {
	if _, err := os.Stat(dataDir + "/" + instancesList); err != nil {
		saveInstances()
	}
	file, err := os.ReadFile(dataDir + "/" + instancesList)
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, &instances)
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		return errNoInstance
	}
	return nil
}

func listInstances() {
	err = readInstances()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Added invidious instances:")
	fmt.Println("---------------------")
	for _, i := range instances {
		fmt.Printf("%s, ping: %.4f, online: %v\n", i.Url, i.Ping, i.Online)
	}
}

func saveInstances() {
	jdata, err := json.Marshal(instances)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(dataDir+"/"+instancesList, jdata, 0755)
	if err != nil {
		panic(err)
	}
}

func restGet(instance, endpoint string, requests map[string]string) (*resty.Response, error) {
	client := resty.New()
	rq := client.R()
	for i, k := range requests {
		rq.SetQueryParam(i, k)
	}
	var resp *resty.Response
	resp, err = rq.Get(instance + endpoint)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("%s", resp.Status())
	}
	header := resp.Header().Values("Content-Type")[0]
	if header != "application/json" {
		return nil, fmt.Errorf("invalid header %s", header)
	}
	return resp, nil
}
