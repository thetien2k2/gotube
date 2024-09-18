# gotube
An app allow user to manage a list of youtube channels, get videos from them and sort videos by some criterias.  
User can play video in mpv by select video from app, able to play next video in list when last video eof.  
App depends on youtube-dl/yt-dlp, mpv and Invidious.

## Build guide
git pull and go build (follow instruction on internet)

## Usage:
Data for app located in user's config directory or user's home directory (when config directory not available).  
Please add some channels at first.

### channel management:
- to add a channel
./gotube addc [channel's url or id]  
example: ./gotube add UCyJnvNM8SX3hiiRdEh6H9vw (get channel's url from https://www.youtube.com/channel/UCyJnvNM8SX3hiiRdEh6H9vw)

- to list added channels  
./gotube lsc 

- to remove a channel
./gotube rmc [channel's url or id]  

### Invidious instance management:
- to add an instance 
./gotube addi [instance's url]

- to list and update instances by responce time
./gotube lsi

- to remove an instance
./gotube rmi [instance's url]

### playing videos
./gotube

App functions:  
- Enter, play video in mpv
- a, toggle sort by publish date
- s, toggle sort by view count
- d, toggle sort by video length
- f, toggle sort by channel name
- q, quit app
- w, update new videos from channels
- e, new video list by search
- o, export list to file playlist.m3u
- r, toggle continuous playing (off by default)
- t, audio only  

When playing video in mpv, press q to quit mpv and come back to playlist

