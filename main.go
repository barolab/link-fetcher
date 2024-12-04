package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kozhurkin/pipers"
	"github.com/urfave/cli/v3"
)

// Filter a slice to keep only unique values
// Taken from https://gist.github.com/johnwesonga/6301924
func unique[T comparable](input []T) []T {
	unique := make([]T, 0, len(input))
	seen := make(map[T]bool, len(input))
	for _, element := range input {
		if !seen[element] {
			unique = append(unique, element)
			seen[element] = true
		}
	}
	return unique
}

// Get an HTML document from the given URL
func getDocument(ctx context.Context, client *http.Client, url *url.URL) (*goquery.Document, error) {
	// Set a timeout of 5s for the GET request
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Make a request and specify that we accept only HTML as a response
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	req.Header.Set("Accept", "text/html")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get '%s', %w", url.String(), err)
	}
	defer res.Body.Close()

	// Make sure the response we get is valid
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get '%s', got response %d (expected a 200)", url.String(), res.StatusCode)
	}

	// Build an HTML document from the response
	// Taken from https://www.zenrows.com/blog/goquery#install-import-goquery
	body, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML from response body, %w", err)
	}

	return body, nil
}

// Transform a list of links into a hash map where the keys are hosts and values are the relative paths
func toHash(links []string) (map[string][]string, error) {
	hash := map[string][]string{}

	for _, link := range links {
		u, err := url.Parse(link)
		if err != nil {
			return hash, fmt.Errorf("link '%s' is not a valid URL, %w", link, err)
		}

		// Build the hash key (which is the URL base domain)
		key := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		// The path is the full URL without the base domain (contain also the fragment and queries if set)
		path := strings.Replace(link, key, "", 1)

		if list, ok := hash[key]; ok {
			hash[key] = append(list, path)
		} else {
			hash[key] = []string{path}
		}
	}

	return hash, nil
}

// Get all the links found in the given URLs
//
// The links are the content of the <a href="...">...</a> attributes
// They are reconstructed in case the href attribute contains only a relative path
// For example '/bar' will become 'https://foo.com/bar'
func getLinksFromURLs(ctx context.Context, urls []string) ([]string, error) {
	// Share the same HTTP client for all the URLs
	client := &http.Client{Transport: &http.Transport{}}

	// Set a timeout of 30s to fetch all the URLs
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Start a routing for each URL, using a waitgroup like library (eg. piper module)
	// Taken from https://stackoverflow.com/questions/59095501/how-to-handle-errors-and-terminate-goroutine-using-waitgroup
	pp := pipers.FromArgs(urls, func(i int, u string) ([]string, error) {
		// Try to parse the URL, making sure it's valid
		ur, err := url.Parse(u)
		if err != nil {
			return nil, fmt.Errorf("argument '%s' is not a valid URL, %w", u, err)
		}

		// Get a GoQuery document from the URL
		// GoQuery supports parsing HTML documents using CSS selectors
		doc, err := getDocument(ctx, client, ur)
		if err != nil {
			return []string{}, err
		}

		// Using a CSS selector to get all the <a ...></a> href attributes
		// Taken from https://www.zenrows.com/blog/goquery#get-links
		links := doc.Find("a").Map(func(i int, a *goquery.Selection) string {
			link, _ := a.Attr("href")

			// Make sure the link is full (in case the href attribute is a relative path)
			if strings.HasPrefix(link, "http") {
				return link
			}

			// If the link starts with a "/" don't append it to the base domain
			if strings.HasPrefix(link, "/") {
				return fmt.Sprintf("%s://%s%s", ur.Scheme, ur.Host, link)
			}

			return fmt.Sprintf("%s://%s/%s", ur.Scheme, ur.Host, link)
		})

		return links, nil
	})

	// Run the pipe (wait for all routine to complete or fail)
	results, err := pp.Context(ctx).Resolve()
	if err != nil {
		return []string{}, err
	}

	// Concatenate all the links found into a single list, and remove duplicates
	return unique(slices.Concat(results...)), nil
}

// Main function whose sole purpose is to run the CLI application
func main() {
	cmd := &cli.Command{

		// Define the flag used to configure the output format
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Value:   "stdout",
				Usage:   "Output format",
				Aliases: []string{"o"},
				Sources: cli.EnvVars("OUTPUT"),
			},
		},
		Name:  "link-fetcher",
		Usage: "list all links of given URLs",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Get all the links found in the given URLs
			links, err := getLinksFromURLs(ctx, cmd.Args().Slice())
			if err != nil {
				return err
			}

			// Print the result based (format depends on the output flag)
			output := cmd.String("output")
			switch output {
			case "stdout":
				for _, link := range links {
					fmt.Println(link)
				}
			case "json":
				// Transform the list of links to a hashmap
				hash, err := toHash(links)
				if err != nil {
					return err
				}
				buf, err := json.Marshal(hash)
				if err != nil {
					return err
				}
				fmt.Println(string(buf))
			default:
				return fmt.Errorf("unkown format %s, supported values are 'json' and 'stdout'", output)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
