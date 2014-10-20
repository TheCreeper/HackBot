package main

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

var (
	ConfigFile string
)

type ClientConfig struct {
	Globals struct {
		Nick        string
		UserName    string
		RealName    string
		Password    string
		CTCPVersion string

		ReconnectIntervalSeconds int
	}

	Proxys []struct {
		Name     string
		Network  string
		Address  string
		Username string
		Password string
		Timeout  time.Duration
	}

	Servers []Server
}

type Server struct {
	AutoConnect bool

	Name   string
	Server string
	UseTLS bool
	Proxy  string

	ProxyNetwork  string
	ProxyAddress  string
	ProxyUsername string
	ProxyPassword string
	ProxyTimeout  time.Duration

	Channels    string
	Nick        string
	UserName    string
	RealName    string
	Password    string
	CTCPVersion string

	ReconnectIntervalSeconds int
	ReconnectMultiplier      int
}

func (cfg *ClientConfig) validate() (err error) {

	var glob = cfg.Globals
	var srv = cfg.Servers
	for i, _ := range cfg.Servers {

		if srv[i].Nick == "" {

			srv[i].Nick = glob.Nick
		}
		if srv[i].UserName == "" {

			srv[i].UserName = glob.UserName
		}
		if srv[i].RealName == "" {

			srv[i].RealName = glob.RealName
		}
		for _, pv := range cfg.Proxys {

			if srv[i].Proxy == pv.Name {

				srv[i].ProxyNetwork = pv.Network
				srv[i].ProxyAddress = pv.Address
				srv[i].ProxyUsername = pv.Username
				srv[i].ProxyPassword = pv.Password
				srv[i].ProxyTimeout = pv.Timeout
			}
		}
		if srv[i].CTCPVersion == "" {

			srv[i].CTCPVersion = glob.CTCPVersion
		}

		if srv[i].ReconnectIntervalSeconds == 0 {

			srv[i].ReconnectIntervalSeconds = glob.ReconnectIntervalSeconds
		}
	}

	return
}

func GetCFG(f string) (cfg ClientConfig, err error) {

	b, err := ioutil.ReadFile(f)
	if err != nil {

		return
	}

	err = json.Unmarshal(b, &cfg)
	if err != nil {

		return
	}

	err = cfg.validate()
	if err != nil {

		return
	}

	return
}
