package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"runtime"
)

func NewServer() (net.Listener, error) {

	return serverListener_unix()
}

func Listen(l net.Listener) error {

	conn, err := l.Accept()
	if err != nil {

		return err
	}
	rpc.ServeConn(conn)

	return nil
}

func serverListener(minPort, maxPort int64) (net.Listener, error) {

	if runtime.GOOS == "windows" {

		return serverListener_tcp(minPort, maxPort)
	}

	return serverListener_unix()
}

func serverListener_tcp(minPort, maxPort int64) (net.Listener, error) {

	for port := minPort; port <= maxPort; port++ {

		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {

			return listener, nil
		}
	}

	return nil, errors.New("Couldn't bind plugin TCP listener")
}

func serverListener_unix() (net.Listener, error) {

	tf, err := ioutil.TempFile("", "irc-plugin")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	// Close the file and remove it because it has to not exist for
	// the domain socket.
	if err := tf.Close(); err != nil {

		return nil, err
	}
	if err := os.Remove(path); err != nil {

		return nil, err
	}

	return net.Listen("unix", path)
}
