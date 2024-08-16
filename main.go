package main

import (
	"flag"
	"fmt"
	"github.com/gocolly/colly/v2"
	"strings"
)

type cssSelectors []string

func (s *cssSelectors) String() string {
	return fmt.Sprintf("%s", strings.Join(*s, ","))
}

func (s *cssSelectors) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var cssFlags cssSelectors

type allowedDomainsSlice []string

func (a *allowedDomainsSlice) String() string {
	return fmt.Sprintf("%s", strings.Join(*a, ","))
}

func (a *allowedDomainsSlice) Set(value string) error {
	*a = append(*a, value)
	return nil
}

var allowedDomains allowedDomainsSlice

func main() {

	flag.Var(&cssFlags, "css", "CSS selector to follow links")

	flag.Var(&allowedDomains, "allowedDomain", "restrict allowed domains")

	urlPaginationPtr := flag.String("urlPaginationPattern", "", "Pattern to follow pagination links")

	flag.Parse()

	urls := flag.Args()

	var c *colly.Collector

	if len(allowedDomains) > 0 {
		c = colly.NewCollector(
			colly.AllowedDomains(allowedDomains...),
			colly.Async(true),
			colly.CacheDir("./web_scraper_cache"),
		)
	} else {
		c = colly.NewCollector(
			colly.Async(true),
			colly.CacheDir("./web_scraper_cache"),
		)
	}

	err := c.Limit(&colly.LimitRule{
		// limit the parallel requests to 4 request at a time
		Parallelism:  4,
		DomainRegexp: ".*",
	})

	if err != nil {
		fmt.Printf("Error setting limit: %s\n", err)
		return
	}

	if len(cssFlags) > 0 {
		for _, cssFlag := range cssFlags {
			c.OnHTML(cssFlag, func(e *colly.HTMLElement) {
				url := e.ChildAttr("a[href]", "href")
				err = c.Visit(url)
				if err != nil {
					fmt.Printf("Error following link %q: %s\n", url, err)
					return
				}
				fmt.Printf("Visted URL: %q\n", url)
			})
		}
	} else {
		fmt.Println("No css flags specified, following all links.")
		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			url := e.Attr("href")
			err = e.Request.Visit(url)
			if err != nil {
				fmt.Printf("Error following link %q: %s\n", url, err)
				return
			}
			fmt.Printf("Visted URL: %q\n", url)
		})
	}

	if *urlPaginationPtr != "" {
		for i := 1; i <= 100; i++ {
			pageURL := fmt.Sprintf(*urlPaginationPtr, i)
			err = c.Visit(pageURL)
			if err != nil {
				fmt.Printf("Error following link %q: %s\n", pageURL, err)
			}
		}
	} else if len(urls) == 0 {
		fmt.Printf("Error No URLs found.\n")
		return
	} else {
		for _, url := range urls {
			err = c.Visit(url)
			if err != nil {
				fmt.Printf("Error visiting URL %q: %s\n", url, err)
				return
			}
			fmt.Printf("Visted URL: %q\n", url)
		}
	}

	c.Wait()

}
