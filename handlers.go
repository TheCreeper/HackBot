package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/TheCreeper/HackBot/ircutil"
	"github.com/TheCreeper/HackBot/responses"
	"github.com/TheCreeper/HackBot/searchquery/crawler"
	"github.com/TheCreeper/HackBot/searchquery/ddg"
	"github.com/sorcix/irc"
)

type HandlerFuncs struct {

	// Expose some information to the handlers
	Name        string
	Server      string
	Channels    string
	Nick        string
	UserName    string
	RealName    string
	Password    string
	CTCPVersion string

	// Client Connection
	ClientConn *ircutil.ClientConn

	// Dialer
	Dial func(network, addr string) (net.Conn, error)
}

const UserAgent = "Mozilla/5.0 (Windows NT 6.1; rv: 24.0) Geck0/20100101 Firefox/24.0 (Tor Browser Bundle)"

func (h *HandlerFuncs) HandleRPLWelcome(m *irc.Message) (err error) {

	// Print MOTD
	log.Printf("%s: %s %s\n", h.Name, m.Command, m.Trailing)

	// Join some channels
	err = h.ClientConn.Join(h.Channels)
	if err != nil {

		return
	}

	return
}

func (h *HandlerFuncs) HandleJoin(m *irc.Message) (err error) {

	// Print Join messages
	log.Printf("%s: %s %s\n", h.Name, m.Command, m.Trailing)
	return
}

func (h *HandlerFuncs) HandlePirvMsg(m *irc.Message) (err error) {

	// Print Private messagess
	log.Printf("%s: %s %s %s\n", h.Name, m.Command, m.Prefix.Name, m.Trailing)

	// Check for portal reference
	if val, ok := responses.Portal[m.Trailing]; ok {

		err = h.ClientConn.PrivMsg(m.Params[0], val)
		if err != nil {

			log.Printf("ircutil.PrivMsg(): %s\n", err)
			return
		}

		return
	}

	// Check for DDG query
	if strings.HasPrefix(m.Trailing, "!ddg") {

		m.Trailing = strings.TrimPrefix(m.Trailing, "!ddg")
		if m.Trailing == "" {

			err = h.ClientConn.PrivMsg(m.Params[0], fmt.Sprintf("%s: %s", m.Prefix.Name, "No Query Specified"))
			if err != nil {

				log.Printf("ircutil.PrivMsg(): %s\n", err)
				return err
			}
			return
		}

		q := &ddg.Client{

			Dial:   h.Dial,
			NoHTML: true,
		}
		_, text, err := q.FeelingLucky(m.Trailing)
		if err != nil {

			log.Printf("ddg.FeelingLucky(): %s\n", err)
		}
		if len(text) < 1 {

			text = "No Results"
		}

		err = h.ClientConn.PrivMsg(m.Params[0], fmt.Sprintf("%s: %s", m.Prefix.Name, text))
		if err != nil {

			log.Printf("ircutil.PrivMsg(): %s\n", err)
			return err
		}
	}

	// Check if message contains URL
	if crawler.IsURL(m.Trailing) {

		if strings.HasPrefix(m.Trailing, "dontcrawl") {

			return
		}

		c := &crawler.Client{
			Dial: h.Dial,
		}
		r, err := c.Crawl(crawler.ExtractUrl(m.Trailing))
		if err != nil {

			log.Printf("crawler.GetTitle(): %s\n", err)
			return nil
		}
		if len(r.Title) > 1 {

			err = h.ClientConn.PrivMsg(m.Params[0], fmt.Sprintf("^ %s", r.Title))
			if err != nil {

				log.Printf("ircutil.PrivMsg(): %s\n", err)
				return err
			}
		}
	}

	return
}

func (h *HandlerFuncs) HandleUnknownCMD(m *irc.Message) (err error) {

	// Print Unknown commands
	log.Printf("%s: %s %s\n", h.Name, m.Command, m.Trailing)
	return
}
