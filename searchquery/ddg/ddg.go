/*
   Documentation on the API: duckduckgo.com/api
*/

package ddg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
)

// Url of the DuckDuckGo API
const UrlApi = "api.duckduckgo.com"

// Result Types
const (
	Article        = "A"
	Disambiguation = "D"
	Category       = "C"
	Name           = "N"
	Exclusive      = "E"
)

type Client struct {
	Dial   func(network, addr string) (net.Conn, error) // Dialer to use
	UrlApi string                                       // Url to duckduckgo api

	UserAgent string // Useragent used in requests

	Pretty             bool // Return pritty json
	NoHTML             bool // Do not include html in querys
	SkipDisambiguation bool // Skip disambiguation
}

func (c *Client) dial(network, addr string) (net.Conn, error) {

	if c.Dial != nil {

		return c.Dial(network, addr)
	}

	dialer := &net.Dialer{

		DualStack: true,
	}

	return dialer.Dial(network, addr)
}

type Result struct {
	Abstract       string // Topic summary (can contain HTML, e.g. italics)
	AbstractText   string // Topic summary (with no HTML)
	AbstractSource string // Name of Abstract source
	AbstractURL    string // Deep link to expanded topic page in AbstractSource
	Image          string // Link to image that goes with Abstract
	Heading        string // Name of topic that goes with Abstract

	Answer     string // Instant answer
	AnswerType string // Type of Answer, e.g. calc, color, digest, info, ip, iploc, phone, pw, rand, regexp, unicode, upc, or zip (see goodies & tech pages for examples).

	Definition       string // Dictionary definition (may differ from Abstract)
	DefinitionSource string // Name of Definition source
	DefinitionURL    string // Deep link to expanded definition page in DefinitionSource

	RelatedTopics []struct { // Array of internal links to related topics associated with Abstract

		Result   string // HTML link(s) to related topic(s)
		FirstURL string // First URL in Result

		Icon struct { // Icon associated with related topic(s)

			URL    string      // URL of icon
			Height interface{} // Height of icon (px)
			Width  interface{} // Width of icon (px)
		}

		Text string // Text from first URL
	}

	Results []struct { // Array of external links associated with Abstract

		Result   string // HTML link(s) to external site(s)
		FirstURL string // First URL in Result

		Icon struct { // Icon associated with related topic(s)

			URL    string      // URL of icon
			Height interface{} // Height of icon (px)
			Width  interface{} // Width of icon (px)
		}

		Text string // Text from FirstURL
	}

	Type     string // Response category, i.e. A (article), D (disambiguation), C (category), N (name), E (exclusive), or nothing.
	Redirect string // !bang redirect URL
}

func (c *Client) Query(searchquery string) (r Result, err error) {

	// If UrlAPI has no value then used UrlAPI const
	if len(c.UrlApi) < 1 {

		c.UrlApi = UrlApi
	}

	// Create the query
	urlF := fmt.Sprintf("https://%s?q=%s&format=json&pretty=%d&no_html=%d&skip_disambig=%d",
		c.UrlApi,
		url.QueryEscape(searchquery),
		boi(c.Pretty),
		boi(c.NoHTML),
		boi(c.SkipDisambiguation))

	httpClient := &http.Client{Transport: &http.Transport{Dial: c.dial}}
	req, err := http.NewRequest("GET", urlF, nil)
	if err != nil {

		return
	}
	if len(c.UserAgent) > 1 {

		req.Header.Add("User-Agent", c.UserAgent)
	}

	// Start the query
	resp, err := httpClient.Do(req)
	if err != nil {

		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		return
	}

	err = json.Unmarshal(body, &r)
	if err != nil {

		return
	}

	return
}

func (c *Client) FeelingLucky(searchuqery string) (typ string, text string, err error) {

	r, err := c.Query(searchuqery)
	if err != nil {

		return
	}

	if r.Type == Disambiguation {

		for _, topic := range r.RelatedTopics {

			if len(topic.Text) < 1 {

				continue
			}
			return r.Type, topic.Text, nil
		}
	}

	if r.Type == Exclusive {

		return r.Type, r.Answer, nil
	}

	return
}
