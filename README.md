# gotubeplaylist
An app allow user to manage a list of youtube channels, get videos from them and sort videos by some criterias.  
User can play video in mpv by select video from app, able to play next video in list when last video eof.  
App depends on youtube-dl/yt-dlp, mpv and Invidious.

## Build guide
git pull and go build (follow instruction on internet)

## Usage:
Data for app located in user's config directory or user's home directory (when config directory not available).  
Please add some channels at first.

### channel management:
App use channel's url or id as idetity.

- to add a channel
./gotubeplaylist add [channel's url or id]  
example: ./gotubeplaylist add UCyJnvNM8SX3hiiRdEh6H9vw (get channel's url from https://www.youtube.com/channel/UCyJnvNM8SX3hiiRdEh6H9vw)

- to list added channels  
./gotubeplaylist list   
./gotubeplaylist ls 

- to delete a channel from list  
./gotubeplaylist delete [channel's url or id]  
./gotubeplaylist rm [channel's url or id]  

### playing videos
./gotubeplaylist

App functions:  
- Enter, play video in mpv
- Ctrl-A, toggle sort by publish date
- Ctrl-S, toggle sort by view count
- Ctrl-D, toggle sort by video length
- Ctrl-F, toggle sort by channel name
- Ctrl-E, export list to file playlist.m3u
- Ctrl-R, toggle continuous playing (off by default)
- Ctrl-U, update videos from channels
- Ctrl-C, exit app

When playing video in mpv, press q to quit mpv and come back to playlist

