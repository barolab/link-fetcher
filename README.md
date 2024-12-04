# Link Fetcher

A Go CLI application to list all links of given URLs

## Running locally

You'll need Go install on your machine (if you're an `asdf` user you can simply run `asdf install`)

A `Makefile` is present for common tasks:

```sh
make build # Build the link-fetcher binary
make fmt   # Format the code
make lint  # Run linters
make test  # Execute the application with https://news.ycombinator.com/ as the main argument
```

## Usage

The application takes one or more URLs to fetch and will list all the links present in the response.
Here's some examples:

```sh
# List all the linkgs of a given URL, display results as pure text with one link per line
$ go run main.go -o stdout https://news.ycombinator.com/
https://news.ycombinator.com
https://news.ycombinator.com/news
https://news.ycombinator.com/newest
https://news.ycombinator.com/front
https://news.ycombinator.com/newcomments
https://news.ycombinator.com/ask
https://news.ycombinator.com/show
https://news.ycombinator.com/jobs
https://news.ycombinator.com/submit
https://news.ycombinator.com/login?goto=news
https://news.ycombinator.com/vote?id=42299863&how=up&goto=news
https://www.map.cv/blog/redbook
...

# Support JSON output where result is hash map where keys are domains and values are paths
$ go run main.go -o json https://news.ycombinator.com/ | jq -r '.'
{
  "https://academic.oup.com": ["/nar/advance-article/doi/10.1093/nar/gkae908/7875984"],
  "https://arxiv.org": ["/abs/2411.19826"],
  "https://christopherferan.com": ["/2021/12/25/kenya-and-the-decline-of-the-worlds-greatest-coffee/"],
  "https://drive.google.com": ["/file/d/1nkHWqYre866xihxN3MnHr5YFzY4gQWDT/view"],
  "https://github.com": ["/KopiasCsaba/open_sound_control_bridge", "/lmnr-ai/flow", "/briangu/klongpy", "/tikimcfee/LookAtThat", "/SeekStorm/SeekStorm", "/HackerNews/API" ],
  ...
}

# Also support multiple URL as arguments
$ go run main.go https://kubernetes.io/ https://kubernetes.io/docs/home/
https://kubernetes.io/
https://kubernetes.io/docs/
https://kubernetes.io/blog/
https://kubernetes.io/training/
https://kubernetes.io/partners/
https://kubernetes.io/community/
https://kubernetes.io/case-studies/
https://kubernetes.io/#
...
```
