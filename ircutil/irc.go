/*
   TODO:
		- Handle errors returned by server
		- Add TLS
*/

package ircutil

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/sorcix/irc"
)

// Errors
var (
	ErrParseMsg   = errors.New("Unable to parse message")
	ErrInvalidMsg = errors.New("Message contains invalid characters")
)

// Some regular expressions
var (
	// Valid message. Make sure it contains no newline chars
	// or the like that could allow for irc injection
	ValidMsg = regexp.MustCompile(`^(.)+$`)
)

type ClientConn struct {
	Dial      func(network, addr string) (net.Conn, error)
	TlsConfig *tls.Config
	Conn      *irc.Conn

	Address  string
	Password string
	Timeout  time.Duration

	Nick       string
	UserName   string
	RealName   string
	OpPassword string
}

func (cc *ClientConn) dial(network, addr string) (net.Conn, error) {

	if cc.Dial != nil {

		return cc.Dial(network, addr)
	}

	d := &net.Dialer{

		Timeout:   cc.Timeout,
		DualStack: true,
	}
	return d.Dial(network, addr)
}

func (cc *ClientConn) Connect(h *Handlers) (err error) {

	conn, err := cc.dial("tcp", cc.Address)
	if err != nil {

		return
	}
	if cc.TlsConfig != nil {

		conn = tls.Client(conn, cc.TlsConfig)
	}
	cc.Conn = irc.NewConn(conn)

	// Run the RegisterConnection handler if ClientConnected not defined
	if h.ClientConnected == nil {

		if err = cc.RegisterClient(); err != nil {

			return
		}
	} else {

		if err = h.ClientConnected(); err != nil {

			return
		}
	}

	// Run the handlers
	if h != nil {

		for {

			err = cc.RunHandlers(h)
			if err != nil {

				return
			}
		}
	}

	return
}

func (cc *ClientConn) Close() error {

	return cc.Conn.Close()
}

func (cc *ClientConn) Disconnect() error {

	err := cc.Quit()
	if err != nil {

		return err
	}

	return cc.Close()
}

func (cc *ClientConn) RegisterClient() (err error) {

	if len(cc.Password) > 1 {

		err = cc.SetPassword()
		if err != nil {

			return
		}
	}

	err = cc.SetNick(cc.Nick)
	if err != nil {

		return
	}

	err = cc.SetUser()
	if err != nil {

		return
	}

	if len(cc.OpPassword) > 1 {

		err = cc.SetOper()
		if err != nil {

			return
		}
	}

	return
}

func (cc *ClientConn) SendRaw(message string) (err error) {

	m := irc.ParseMessage(message)
	if m == nil {

		return ErrParseMsg
	}

	err = cc.Conn.Encode(m)
	if err != nil {

		return
	}

	return
}

func (cc *ClientConn) PingPong(m *irc.Message) error {

	if m.Command != irc.PING {

		return nil
	}

	return cc.Pong(m.Trailing)
}

type Handlers struct {
	ClientConnected func() error
	RPLWelcome      func(*irc.Message) error
	PingPong        func(*irc.Message) error
	Join            func(*irc.Message) error
	PrivMsg         func(*irc.Message) error
	UnknownCMD      func(*irc.Message) error
}

func (cc *ClientConn) RunHandlers(h *Handlers) (err error) {

	message, err := cc.Conn.Decode()
	if err != nil {

		return
	}

	switch message.Command {

	case irc.RPL_WELCOME:

		if h.RPLWelcome == nil {

			return
		}

		return h.RPLWelcome(message)

	case irc.JOIN:

		if h.Join == nil {

			return
		}

		return h.Join(message)

	case irc.PING:

		if h.PingPong == nil {

			return cc.PingPong(message)
		}

		return h.PingPong(message)

	case irc.PRIVMSG:

		if h.PrivMsg == nil {

			return
		}

		return h.PrivMsg(message)

	default:

		if h.UnknownCMD == nil {

			return
		}

		return h.UnknownCMD(message)
	}

	return
}

/*
   @Connection Registration
   RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.1
*/

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.1.1
func (cc *ClientConn) SetPassword() error {

	return cc.SendRaw(fmt.Sprintf("%s %s\r\n", irc.PASS, cc.Password))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.1.2
func (cc *ClientConn) SetNick(nick string) error {

	// Sanitise
	if !(ValidMsg.MatchString(nick)) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %s\r\n", irc.NICK, nick))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.1.3
func (cc *ClientConn) SetUser() error {

	return cc.SendRaw(fmt.Sprintf("%s %s * * :%s\r\n", irc.USER, cc.UserName, cc.RealName))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.1.5
func (cc *ClientConn) SetOper() error {

	return cc.SendRaw(fmt.Sprintf("%s %s %s\r\n", irc.OPER, cc.UserName, cc.OpPassword))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.1.6
func (cc *ClientConn) Quit() error {

	return cc.SendRaw(fmt.Sprintf("%s %s\r\n", irc.QUIT, "Bye Bye"))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.1.6
func (cc *ClientConn) QuitM(message string) error {

	// Sanitise
	if !(ValidMsg.MatchString(message)) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %s\r\n", irc.QUIT, message))
}

/*
   @Channel Operations
   RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.2
*/

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.2.1
func (cc *ClientConn) Join(channels string) error {

	// Sanitise
	if !(ValidMsg.MatchString(channels)) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %s\r\n", irc.JOIN, channels))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.2.2
func (cc *ClientConn) Part(channel string) error {

	// Sanitise
	if !(ValidMsg.MatchString(channel)) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %s\r\n", irc.PART, channel))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.2.3
func (cc *ClientConn) Mode(target, mode string) error {

	// Sanitise
	if !ValidMsg.MatchString(target) && !ValidMsg.MatchString(mode) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %s %s\r\n", irc.MODE, target, mode))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.2.4
func (cc *ClientConn) Topic(target, topic string) error {

	// Sanitise
	if !ValidMsg.MatchString(target) && !ValidMsg.MatchString(topic) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %s %s\r\n", irc.TOPIC, target, topic))
}

/*
   @Sending messages
   tools.ietf.org/html/rfc1459.html#section-4.4
*/

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.4.1
func (cc *ClientConn) PrivMsg(target, message string) error {

	// Sanitise
	if !ValidMsg.MatchString(target) && !ValidMsg.MatchString(message) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %s :%s\r\n", irc.PRIVMSG, target, message))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.4.1
func (cc *ClientConn) Notice(target, message string) error {

	// Sanitise
	if !ValidMsg.MatchString(target) && !ValidMsg.MatchString(message) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %s :%s\r\n", irc.NOTICE, target, message))
}

/*
   @Miscellaneous messages
   tools.ietf.org/html/rfc1459.html#section-4.6
*/

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.6.2
func (cc *ClientConn) Ping(message string) error {

	// Sanitise
	if !(ValidMsg.MatchString(message)) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %d\r\n", irc.PING, message))
}

// RFC 1459 details: tools.ietf.org/html/rfc1459#section-4.6.3
func (cc *ClientConn) Pong(message string) error {

	// Sanitise
	if !ValidMsg.MatchString(message) {

		return ErrInvalidMsg
	}

	return cc.SendRaw(fmt.Sprintf("%s %s\r\n", irc.PONG, message))
}
