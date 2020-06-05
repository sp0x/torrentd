[![sp0x](https://circleci.com/gh/sp0x/rutracker-rss.svg?style=shield)](https://circleci.com/gh/sp0x/rutracker-rss)
[![codecov](https://codecov.io/gh/sp0x/rutracker-rss/branch/master/graph/badge.svg)](https://codecov.io/gh/sp0x/rutracker-rss)
[![Go Report Card](https://goreportcard.com/badge/github.com/sp0x/rutracker-rss)](https://goreportcard.com/report/github.com/sp0x/rutracker-rss)

# Torrentd
This project aims to make torrent searching easier.  
You can use it as a torrent indexer to connect servers like Sonarr, Radarr or others.    
It gathers additional information about each torrent and enriches its data.   
You can also use it to track torrent sites, if you want to mirror a tracker, or a record on torrents.

## Torrent Tracker definitions
You can define your torrent trackers in these directories:
- ~/.torrentd/definitions
- <currentDirectory>/definitions

This project also carries it's embedded definitions with which it was built.  
The definition that's loaded is the latest one.

## Storage
This project stores all the discovered torrents in a SQLite database in `./db/main.db`.

## Caching
By default the server caches the following data:
- Connectivity checks (LRU with Timeout)
- Search results (LRU with Timeout)

