# gotubeplaylist
Create list of youtube videos by scalping videos from subcripbed channels and sort by some criterials

## Build guide
git pull and go build (follow instruction on internet)

## Usage:
First, add some channels to subscripbed list (file channels.json), by using this command (for each channel)

./gotubeplaylist add [channel ID or Handler or URL]

Open app by enter ./gotubeplaylist in terminal.

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

