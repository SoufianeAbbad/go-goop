package main

import (
	"crypto/tls"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Result struct {
	Url   string
	Title string
}
func getClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	re := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &http.Client{
		Transport:     tr,
		CheckRedirect: re,
		Timeout:       time.Second * 20,
	}
}

func search(c *http.Client, query string, cookie string, page int, full bool) []Result {
	var searchResults []Result

	filter := 0
	if full != true {
		filter = 1
	}
	offset := page * 100

	// Encode google parameters
	u1, err := url.Parse("https://google.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Url parse error")
		return searchResults
	}
	u1.Path += "/search"
	p1 := url.Values{}
	p1.Add("q", query)
	p1.Add("start", strconv.Itoa(offset))
	p1.Add("filter", strconv.Itoa(filter))
	p1.Add("num", "100")
	u1.RawQuery = p1.Encode()

	// Encode facebook parameters
	u2, err := url.Parse("https://developers.facebook.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Url parse error")
		return searchResults
	}
	u2.Path += "/tools/debug/echo/"
	p2 := url.Values{}
	p2.Add("q", u1.String())
	u2.RawQuery = p2.Encode()

	// Make a request
	req, err := http.NewRequest("GET", u2.String(), nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return searchResults
	}
	req.Header.Set("Host", "developers.facebook.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:68.0) Gecko/20100101 Firefox/68.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "deflate")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("TE", "Trailers")

	resp, err := c.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error requesting %s: %s\n", u2, err)
		return searchResults
	}
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error when read response query: %s\n", err.Error())
		return searchResults
	}

	// Response double encoded
	decodedRaw := html.UnescapeString(html.UnescapeString(string(raw)))

	// Use regex to get urls
	r := regexp.MustCompile(`<a href="/url\?q=(.+?)&sa=.*"><div class="[^"]+">(.*?)</div>`)
	matches := r.FindAllStringSubmatch(decodedRaw, -1)

	for _, v := range matches {
		r := Result{
			Url:   v[1],
			Title: v[2],
		}
		searchResults = append(searchResults, r)
	}
	return searchResults
}
