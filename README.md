# Link Fetcher

A Go CLI application to list all links of given URLs

## Running locally

You'll need [Dagger](https://dagger.io/) install on your machine (if you're an `asdf` user you can simply run `asdf install`)

A `Makefile` is present for common tasks:

```sh
make build # Build the link-fetcher Docker image and publish it on ttl.sh
make fmt   # Format the code
make lint  # Run linters
make test  # Execute the application with https://news.ycombinator.com/ as the main argument
```

## Usage

The application takes one or more URLs to fetch and will list all the links present in the response.
There's a couple of options, for example to customize the output format. An help is available using:

```sh
$ ./link-fetcher --help
NAME:
   link-fetcher - list all links of given URLs

USAGE:
   link-fetcher [global options]

GLOBAL OPTIONS:
   --output value, -o value  Output format (default: "stdout") [$OUTPUT]
   --sleep, -s               Sleep endlessly after printing output (default: false) [$SLEEP]
   --help, -h                show help
```

Here's some examples of what the CLI can do:

```sh
# List all the linkgs of a given URL, display results as pure text with one link per line
$ go run main.go -o stdout https://news.ycombinator.com/
https://news.ycombinator.com
https://news.ycombinator.com/news
https://news.ycombinator.com/newest
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
  "https://github.com": ["/KopiasCsaba/open_sound_control_bridge", "/lmnr-ai/flow", "/briangu/klongpy" ],
  ...
}

# Also support multiple URL as arguments
$ go run main.go https://kubernetes.io/ https://kubernetes.io/docs/home/
https://kubernetes.io/
https://kubernetes.io/docs/
https://kubernetes.io/community/
https://kubernetes.io/case-studies/
...
```

## Docker

The Link Fetcher CLI is also available as a Docker image, which be run like this:

```sh
$ docker run ttl.sh/link-fetcher:latest https://kubernetes.io/docs/home/
https://kubernetes.io/
https://kubernetes.io/docs/
https://kubernetes.io/blog/
https://kubernetes.io/training/
...
```

The build of the image is done using [Dagger](https://dagger.io/). You can use dagger directly using `dagger call build` or use the Makefile overlay `make build`.

A security scan is also available through Dagger. Here's quick preview:

```sh
$ dagger call scan --image=ttl.sh/link-fetcher:latest
ttl.sh/link-fetcher:latest@sha256:8d0052d6659937e83c045473201fd12764a1ef3b7d0ae00fdaee91f4516cc99d (alpine 3.20.3)
==================================================================================================================
Total: 2 (UNKNOWN: 0, LOW: 2, MEDIUM: 0, HIGH: 0, CRITICAL: 0)

┌────────────┬───────────────┬──────────┬────────┬───────────────────┬───────────────┬───────────────────────────────────────────────────────────┐
│  Library   │ Vulnerability │ Severity │ Status │ Installed Version │ Fixed Version │                           Title                           │
├────────────┼───────────────┼──────────┼────────┼───────────────────┼───────────────┼───────────────────────────────────────────────────────────┤
│ libcrypto3 │ CVE-2024-9143 │ LOW      │ fixed  │ 3.3.2-r0          │ 3.3.2-r1      │ openssl: Low-level invalid GF(2^m) parameters lead to OOB │
│            │               │          │        │                   │               │ memory access                                             │
│            │               │          │        │                   │               │ https://avd.aquasec.com/nvd/cve-2024-9143                 │
├────────────┤               │          │        │                   │               │                                                           │
│ libssl3    │               │          │        │                   │               │                                                           │
│            │               │          │        │                   │               │                                                           │
│            │               │          │        │                   │               │                                                           │
└────────────┴───────────────┴──────────┴────────┴───────────────────┴───────────────┴───────────────────────────────────────────────────────────┘
```

> As seen in the report above, the SSL & Crypto libraries are out of date.
> They contain a low level CVE, which given the description on the links seems to impact only NodeJS on Ubuntu
> Anyway, keeping libs up to date is better which is why we run `apk update && apk upgrade` during the Docker build

## Kubernetes

Link Fetcher can also be deployed on Kubernetes using the manifest availabe in [kubernetes/deployment.yaml](/kubernetes/deployment.yaml).
To make it easier to deploy and validate, some dagger functions are there to help, here's some helpful commands:

```sh
# Start a local K3S cluster, this will run forever, you'll have to Ctrl+C to stop the cluster
$ dagger call -m "github.com/marcosnils/daggerverse/k3s@k3s/v0.1.7" --name=test server up

# In another terminal, you can get the cluster KUBECONFIG file on your host with
$ dagger call -m "github.com/marcosnils/daggerverse/k3s@k3s/v0.1.7" --name=test config export --path=/tmp/kubeconfig

# You can use that module to run kubectl commands; like checking namespaces are present
$ dagger call -m "github.com/marcosnils/daggerverse/k3s@k3s/v0.1.7" --name=test kubectl --args="get ns" stdout

# You can also spin up a K9S terminal on the K3S cluster for troubleshooting
$ dagger call -m "github.com/marcosnils/daggerverse/k3s@k3s/v0.1.7" --name=test kns terminal

# To deploy the application, simply call our own Dagger function with the K3S cluster KUBECONFIG file
$ dagger call deploy --kubeconfig=/tmp/kubeconfig
deployment.apps/link-fetcher created

# You can check that the Deployment is runnning
$ dagger call status --kubeconfig=/tmp/kubeconfig
Waiting for deployment "link-fetcher" rollout to finish: 0 of 1 updated replicas are available...
deployment "link-fetcher" successfully rolled out

# And finally you can retrive the Link Fetcher logs, and validate the output is valid
$ dagger call logs --kubeconfig=/tmp/kubeconfig
{"http://timhulsizer.com":["/cwords/chonk.html"], ...}
$ dagger call validate --kubeconfig=/tmp/kubeconfig
Found 29 results, validation succeeded
```


To test a full pipeline locally, you can also use the `dagger call integration-test` command.
Internally this command will:
- Build the Docker image and publish it to ttl.sh
- Start a [K3S cluster](./dagger/main.go#L214)
- Deploy the Link Fetcher [K8S manifest on it](./dagger/main.go#L223)
- Validate the [Deployment is working](./dagger/main.go#159)
    - The validation first wait for the Deployment rollout to complete
    - It gets the Deployment logs
    - Parse it as JSON
    - Count the amount of entries in the hash map
    - Fail if there's no entries at all

Here's a sneak peak:

```sh
$ dagger call integration-test
✔ connect 0.1s
✔ loading module 3.2s
✔ parsing command line arguments 0.0s
✔ linkFetcher: LinkFetcher! 0.0s
✔ LinkFetcher.integrationTest: String! 40.1s
! call function "IntegrationTest": context canceled
  ✔ Container.from(address: "golang:latest"): Container! 0.4s
  ✔ Container.from(address: "alpine"): Container! 0.2s
  ✔ Container.publish(address: "ttl.sh/link-fetcher:latest"): String! 1.3s
  ✔ k3S(name: "test"): K3S! 0.5s
  ✔ K3S.server: Service! 0.3s
  ✔ Service.start: ServiceID! 0.3s
  ✔ K3S.config: File! 0.4s
  ✔ Container.from(address: "bitnami/kubectl"): Container! 0.2s
  ✔ Container.withExec(args: ["sh", "-c", "kubectl apply -f /src/kubernetes/deployment.yaml"]): Container! 0.2s
  ✔ Container.stdout: String! 0.3s
  ✔ Container.withExec(args: ["sh", "-c", "kubectl rollout status deploy link-fetcher --timeout=60s"]): Container! 33.2s
  ✔ Container.stdout: String! 33.2s
  ✔ Container.withExec(args: ["sh", "-c", "kubectl logs deploy/link-fetcher"]): Container! 0.2s
  ✔ Container.stdout: String! 0.2s

- build:
ttl.sh/link-fetcher:latest@sha256:ef7acb2c62ca068457f18e6a3f53679e092724a64128e79c768dc7e4ee9602ff

- deploy:
deployment.apps/link-fetcher created

- validate:
Found 27 results, validation succeeded
```

## Playground

I wanted to test if it was possible to extract the base domain name of the URLs printed by Link Fetcher.
You can find the implementation in the [/playground/sanitize.sh](/playground/sanitize.sh) script.

Here's a quick overview of how it works:

```sh
$ cat ./playground/sample.txt 
http://tiktok.com
https://ads.faceBook.com.
https://sub.ads.faCebook.com
api.tiktok.com
Google.com.
aws.amazon.com

$ ./playground/sanitize.sh 
- First attempt with grep:
amazon.com
facebook.com
google.com
tiktok.com

- Second attempt with awk:
amazon.com
facebook.com
google.com
tiktok.com
```

Sadly, the current implementation works only for domain names ending the `.com` suffix, so that's somthing to improve in the futur.
