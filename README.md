[![sp0x](https://circleci.com/gh/sp0x/torrentd.svg?style=shield)](https://circleci.com/gh/sp0x/torrentd)
[![codecov](https://codecov.io/gh/sp0x/torrentd/branch/master/graph/badge.svg)](https://codecov.io/gh/sp0x/torrentd)
[![Go Report Card](https://goreportcard.com/badge/github.com/sp0x/torrentd)](https://goreportcard.com/report/github.com/sp0x/torrentd)

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
All definitions are yml files.

## Storage
You'll also need to store your results somehow. Depending on the way you run this project there are a few ways you could do that.  
The supported databases are:
 - SQLite
 - BoltDB
 - Firebase

## Configuration
The configuration is stored in: ~/.torrentd/torrentd.yml   
Here's a brief overview of what you can configure:
```yaml
# The key for accessing the API.
api_key: hsreth45hgertdf
# Places where index definitions are stored.
definition:
  dirs:
  - ./definitions
  - ~/.torrentd/definitions
# The port on which the API runs.
port: 5000
# Whether to print more logs.
verbose: false

# Index config:
indexers:
    # We'll configure the zamunda index
    zamunda:
      username: myusername
      password: g43ewef
      #To use the the login creds in the index definition you just need to use them as a template in the login block.
      #Like this:
      #login:
      #  path: takelogin.php
      #  method: post
      #  inputs:
      #    username: "{{ .Config.username }}"
      #    password: "{{ .Config.password }}" 
```

## Index definitions
### Login
Login is described with the following block
```yaml
login:
  # Path that's used for the request
  path: takelogin.php
  # The method can be:
  # post - a post request is sent using the form content-type
  # form - similar as `post` but a specific form can be filled in
  # cookie - a special cookie is set to act as a login session
  method: post
  # A selector can be used here to match a specific form in the page, that should be filled in
  # This can be used only with the `form` login method
  form: #my-form-id 
  # The data that will be sent. Config patterns can be used here using {{ .Config.<field-name> }}
  inputs:
    username: "{{ .Config.username }}"
    password: "{{ .Config.password }}"
  # The error block describes where to look for login errors
  error:
    selector: d.embedded:has(h2:contains("failed"))
  # If the selector has any matches this means that you're logged in
  test:
    selector: a[href="/logout.php"]

```

## Caching
By default, the server caches the following data:
- Connectivity checks (LRU with Timeout)
- Search results (LRU with Timeout)
- Last errors for each loaded index (LRU with 2 days TTL)

