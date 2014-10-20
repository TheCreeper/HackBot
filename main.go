package main

import (
	"flag"
	"io"
	"log"
	"sync"
	"time"

	"github.com/TheCreeper/HackBot/ircutil"
)

func (cfg *ClientConfig) LaunchClient(wg *sync.WaitGroup, srv Server) {

	// Setup the proxy connection
	conn := &ConnHandler{

		ProxyNetwork:  srv.ProxyNetwork,
		ProxyAddress:  srv.ProxyAddress,
		ProxyUsername: srv.ProxyUsername,
		ProxyPassword: srv.ProxyPassword,
		Timeout:       srv.ProxyTimeout,
	}

	// Setup the irc client connection
	cc := &ircutil.ClientConn{

		Address:  srv.Server,
		Nick:     srv.Nick,
		UserName: srv.UserName,
		RealName: srv.RealName,
		Dial:     conn.HandleConnection,
	}

	// Pass some vars to the handlers
	h := &HandlerFuncs{

		Name:        srv.Name,
		Server:      srv.Server,
		Channels:    srv.Channels,
		Nick:        srv.Nick,
		UserName:    srv.UserName,
		RealName:    srv.RealName,
		Password:    srv.Password,
		CTCPVersion: srv.CTCPVersion,

		ClientConn: cc,
		Dial:       conn.HandleConnection,
	}

	// Setup the handlers
	handlers := &ircutil.Handlers{

		RPLWelcome: h.HandleRPLWelcome,
		Join:       h.HandleJoin,
		PrivMsg:    h.HandlePirvMsg,
		UnknownCMD: h.HandleUnknownCMD,
	}

	// Execute main loop
	for {

		time.Sleep(time.Duration(srv.ReconnectIntervalSeconds) * time.Second)

		err := cc.Connect(handlers)
		if err == io.EOF {

			cc.Close()
		}
		if err != nil {

			log.Printf("irc.Connect(): %s\n", err)
			continue
		}
		defer cc.Close()
	}

	wg.Done()
}

func init() {

	flag.StringVar(&ConfigFile, "config", "./config.json", "The configuration file location")
	flag.Parse()
}

func main() {

	var wg sync.WaitGroup

	cfg, err := GetCFG(ConfigFile)
	if err != nil {

		log.Fatal(err)
	}

	for _, v := range cfg.Servers {

		if v.AutoConnect {

			wg.Add(1)
			go cfg.LaunchClient(&wg, v)
		}
	}

	wg.Wait()
}
