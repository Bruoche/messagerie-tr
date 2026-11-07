package config

import (
	"net/url"
	"strconv"
)

type servConfig struct {
	Host   string
	ID     string
	Domain string
}

var Server servConfig
var Debug bool

func Set(host, id string, port int, debug bool) {
	Debug = debug
	SetServ(host, id, port)
}

func SetServ(host, id string, port int) {
	domain := "localhost:" + strconv.Itoa(port)
	if host == "" {
		host = "http://" + domain
	}
	if id == "" {
		u, err := url.Parse(host)
		if err != nil {
			id = domain
		} else {
			id = u.Hostname()
		}
	}
	Server = servConfig{
		Host:   host,
		ID:     id,
		Domain: domain,
	}
}
