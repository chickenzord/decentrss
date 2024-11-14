# DecentRSS

(supposed to be) decentralized RSS proxy

> This project is still in initial/PoC stage, there might be a lot of backward-incompatible changes ahead.

## Background

- I want to RSS-ify as much as my content consumption as possible
- Using RSS-bridge or RSSHub (I also own private instance with custom routes) often get blocked by anti-crawler mechanism (e.g. Cloudflare)
- My Miniflux instance also often get blocked when directly using the website's feeds (e.g. bad VM's IP reputation)
- I need more ways to add contents to the feed reader to circumvent the various blocking

## Diagram

![DecentRSS diagram](docs/diagram-01.png?raw=true "Title")

## Running example

```yml
services:
  decentrss:
    image: ghcr.io/chickenzord/decentrss:master
    container_name: decentrss
    restart: on-failure
    user: 1000:1000
    environment:
      DATA_DIR: /data
    ports:
    - 8081:8080
    volumes:
    - /somewhere-in-the-host/decentrss-data:/data
```

## Usage

### Adding feed items manually via API

```sh
curl https://example.org/feed | curl -XPOST -d @- http://localhost:8080/feed
```

- `/feed` endpoint supports Atom/RSS/JSONfeed payload
- The feed and its items will be saved identified by the URL (i.e. `https://example.org/feed`)

## Using in feed readers

In the example above, you can use this URL in the feed reader:

```
http://localhost:8080/feed?url=https://example.org/feed

```

TODO: feed configurations (e.g. max items to return)

## Automatic feed crawling (not implemented yet)

GET-ing the URL above will automatically fetch and cache the feed contents in DecentRSS storage before serving the response.

## ~~Syncing with~~ Connecting to another DecentRSS instances (not implemented yet)

DecentRSS can get the content from other registered DecentRSS instances as fallback when blocked by the upstream feed.
