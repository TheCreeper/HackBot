/*
	Some examples:
		- godoc.org/github.com/bashtian/htmlparse
		- github.com/madelfio/grid/html2tab
		- github.com/crackcomm/arts
		- code.google.com/p/go.net/html
		- github.com/moovweb/gokogiri
		- golang-examples.tumblr.com/post/47426518779/parse-html
	Notes:
		Regex Url: http://[-A-Za-z0-9+&@#/%?=~_()|!:,.;]*[-A-Za-z0-9+&@#/%=~_()]
*/

package crawler

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"

	//"code.google.com/p/go.net/html"
)

// Errors
var (
	ErrMimeType    = errors.New("MIME type not supported")
	ErrInvalidChar = errors.New("Crawler result contains invalid characters")
)

// Default useragent
const UserAgent = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

// Some regular expressions
var (
	// Valid url
	ValidUrl = regexp.MustCompile(`https?://[-A-Za-z0-9+&@#/%?=~_()|!:,.;]*`)

	// Valid Page Title
	ValidPageTitle = regexp.MustCompile(`^(.)+$`)
)

// Map of allowed mime types
var AllowedMimeTypes = map[string]bool{

	"text/html; charset=utf-8": true,
}

// Some helper funcs
func IsURL(s string) bool {

	return ValidUrl.MatchString(s)
}

func ExtractUrl(s string) string {

	return ValidUrl.FindString(s)
}

// Client structure
type Client struct {

	// Dialer used for requests
	Dial func(network, addr string) (net.Conn, error)

	// The useragent used in requests
	UserAgent string

	// UserName and PassWord for authentication with the WebServer
	UserName string
	PassWord string
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

// Store some information about the page
type CrawlResult struct {
	Title string `xml:"head>title"` // Title of the page
	//Desc string 'xml: "head>meta"' // Description of the page
	Size int // Size of webpage
}

func (c *Client) Crawl(urlF string) (r *CrawlResult, err error) {

	httpClient := &http.Client{

		Transport: &http.Transport{Dial: c.dial},
	}

	req, err := http.NewRequest("GET", urlF, nil)
	if err != nil {

		return
	}
	if len(c.UserAgent) < 1 {

		c.UserAgent = UserAgent
	}
	req.Header.Add("User-Agent", c.UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {

		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		return
	}

	mime := http.DetectContentType(body)
	_, ok := AllowedMimeTypes[mime]
	if !(ok) {

		return nil, ErrMimeType
	}

	autoclose := []string{"script", "style", "meta"}
	for _, v := range xml.HTMLAutoClose {

		autoclose = append(autoclose, v)
	}

	// Populate the crawler result
	d := xml.NewDecoder(bytes.NewBuffer(body))
	d.Strict = false
	d.AutoClose = autoclose
	d.Entity = xml.HTMLEntity
	err = d.Decode(&r)
	if err != nil {

		return
	}

	r.Size = len(body)

	// Sanitise
	if !ValidPageTitle.MatchString(r.Title) {

		return nil, ErrInvalidChar
	}

	/*
		d := html.NewTokenizer(resp.Body)
		for {

			tokenType := d.Next()

			// token type
			token := d.Token()
			switch tokenType {
			case html.StartTagToken: // <tag>

				// type Token struct {
				//     Type     TokenType
				//     DataAtom atom.Atom
				//     Data     string
				//     Attr     []Attribute
				// }
				//
				// type Attribute struct {
				//     Namespace, Key, Val string
				// }
			case html.TextToken: // text between start and end tag
				fmt.Printf("[Found] %v [Node] %v\n", token.Data, token.Attr)
			case html.EndTagToken: // </tag>
			case html.SelfClosingTagToken: // <tag/>
			}
		}
	*/

	return
}

// Based on list from MDN's HTML5 element list
// https://developer.mozilla.org/en-US/docs/Web/Guide/HTML/HTML5/HTML5_element_list
var acceptableElements = map[string]bool{
	// Root element
	// "html": true,

	// Document metadata
	// "head":  true,
	// "title": true,
	// "base":  true,
	// "link":  true,
	// "meta":  true,
	// "style": true,

	// Scripting
	"noscript": true,
	// "script":   true,

	// Sections
	// "body":    true,
	"section": true,
	"nav":     true,
	"article": true,
	"aside":   true,
	"h1":      true,
	"h2":      true,
	"h3":      true,
	"h4":      true,
	"h5":      true,
	"h6":      true,
	"header":  true,
	"footer":  true,
	"address": true,
	"main":    true,

	// Grouping content
	"p":          true,
	"hr":         true,
	"pre":        true,
	"blockquote": true,
	"ol":         true,
	"ul":         true,
	"li":         true,
	"dl":         true,
	"dt":         true,
	"dd":         true,
	"figure":     true,
	"figcaption": true,
	"div":        true,

	// Text-level semantics
	"a":      true,
	"em":     true,
	"strong": true,
	"small":  true,
	"s":      true,
	"cite":   true,
	"q":      true,
	"dfn":    true,
	"abbr":   true,
	"data":   true,
	"time":   true,
	"code":   true,
	"var":    true,
	"samp":   true,
	"kbd":    true,
	"sub":    true,
	"sup":    true,
	"i":      true,
	"b":      true,
	"u":      true,
	"mark":   true,
	"ruby":   true,
	"rt":     true,
	"rp":     true,
	"bdi":    true,
	"bdo":    true,
	"span":   true,
	"br":     true,
	"wbr":    true,

	// Edits
	"ins": true,
	"del": true,

	// Embedded content
	"img":    true,
	"iframe": true,
	"embed":  true,
	"object": true,
	"param":  true,
	"video":  true,
	"audio":  true,
	"source": true,
	"track":  true,
	"canvas": true,
	"map":    true,
	"area":   true,
	"svg":    true,
	"math":   true,

	// Tabular data
	"table":    true,
	"caption":  true,
	"colgroup": true,
	"col":      true,
	"tbody":    true,
	"thead":    true,
	"tfoot":    true,
	"tr":       true,
	"td":       true,
	"th":       true,

	// Forms
	"form":     true,
	"fieldset": true,
	"legend":   true,
	"label":    true,
	"input":    true,
	"button":   true,
	"select":   true,
	"datalist": true,
	"optgroup": true,
	"option":   true,
	"textarea": true,
	"keygen":   true,
	"output":   true,
	"progress": true,
	"meter":    true,

	// Interactive elements
	// "details":  true,
	// "summary":  true,
	// "menuitem": true,
	// "menu":     true,
}

var unacceptableElementsWithEndTag = map[string]bool{
	"script": true,
	"applet": true,
	"style":  true,
}

// Based on list from MDN's HTML attribute reference
// https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes
var acceptableAttributes = map[string]bool{
	"accept":         true,
	"accept-charset": true,
	// "accesskey":       true,
	"action":       true,
	"align":        true,
	"alt":          true,
	"async":        true,
	"autocomplete": true,
	// "autofocus":       true,
	// "autoplay":        true,
	"bgcolor":         true,
	"border":          true,
	"buffered":        true,
	"challenge":       true,
	"charset":         true,
	"checked":         true,
	"cite":            true,
	"class":           true,
	"code":            true,
	"codebase":        true,
	"color":           true,
	"cols":            true,
	"colspan":         true,
	"content":         true,
	"contenteditable": true,
	"contextmenu":     true,
	"controls":        true,
	"coords":          true,
	"data":            true,
	"data-custom":     true,
	"datetime":        true,
	"default":         true,
	"defer":           true,
	"dir":             true,
	"dirname":         true,
	"disabled":        true,
	"download":        true,
	"draggable":       true,
	"dropzone":        true,
	"enctype":         true,
	"for":             true,
	"form":            true,
	"headers":         true,
	"height":          true,
	"hidden":          true,
	"high":            true,
	"href":            true,
	"hreflang":        true,
	"http-equiv":      true,
	"icon":            true,
	"id":              true,
	"ismap":           true,
	"itemprop":        true,
	"keytype":         true,
	"kind":            true,
	"label":           true,
	"lang":            true,
	"language":        true,
	"list":            true,
	"loop":            true,
	"low":             true,
	"manifest":        true,
	"max":             true,
	"maxlength":       true,
	"media":           true,
	"method":          true,
	"min":             true,
	"multiple":        true,
	"name":            true,
	"novalidate":      true,
	"open":            true,
	"optimum":         true,
	"pattern":         true,
	"ping":            true,
	"placeholder":     true,
	"poster":          true,
	// "preload":         true,
	"pubdate":    true,
	"radiogroup": true,
	"readonly":   true,
	"rel":        true,
	"required":   true,
	"reversed":   true,
	"rows":       true,
	"rowspan":    true,
	"sandbox":    true,
	"spellcheck": true,
	"scope":      true,
	// "scoped":          true,
	// "seamless":        true,
	"selected": true,
	"shape":    true,
	"size":     true,
	"sizes":    true,
	"span":     true,
	"src":      true,
	// "srcdoc":          true,
	"srclang": true,
	"start":   true,
	// "step":            true,
	"style":   true,
	"summary": true,
	// "tabindex":        true,
	// "target":          true,
	"title": true,
	"type":  true,
	// "usemap":          true,
	"value": true,
	"width": true,
	// "wrap":            true,

	// Older HTML attributes
	// http://www.w3.org/TR/html5-diff/#obsolete-attributes
	"alink":        true,
	"background":   true,
	"cellpadding":  true,
	"cellspacing":  true,
	"char":         true,
	"clear":        true,
	"compact":      true,
	"frameborder":  true,
	"frame":        true,
	"hspace":       true,
	"marginheight": true,
	"noshade":      true,
	"nowrap":       true,
	"rules":        true,
	"scrolling":    true,
	"valign":       true,
}

// Based on list from Wikipedia's URI scheme
// http://en.wikipedia.org/wiki/URI_scheme
var acceptableUriSchemes = map[string]bool{
	"aim":      true,
	"apt":      true,
	"bitcoin":  true,
	"callto":   true,
	"cvs":      true,
	"facetime": true,
	"feed":     true,
	"ftp":      true,
	"git":      true,
	"gopher":   true,
	"gtalk":    true,
	"http":     true,
	"https":    true,
	"imap":     true,
	"irc":      true,
	"itms":     true,
	"jabber":   true,
	"magnet":   true,
	"mailto":   true,
	"mms":      true,
	"msnim":    true,
	"news":     true,
	"nntp":     true,
	"rtmp":     true,
	"rtsp":     true,
	"sftp":     true,
	"skype":    true,
	"svn":      true,
	"ymsgr":    true,
}
