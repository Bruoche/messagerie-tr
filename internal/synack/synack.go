package synack

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Bruoche/messagerie-tr/internal/request"
	"github.com/Bruoche/messagerie-tr/pkg/ext"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/gin-gonic/gin"
)

type connexionStatus int

const (
	NAN connexionStatus = iota
	SYN
	ACK
)

var hostsToSub = make(chan string, 100)
var hostsToPub = make(chan string, 100)
var connections = map[string]connexionStatus{}

const HeaderHost = "Origin"

func Start() {
	go SendSynAck(hostsToSub)
	go SendAck(hostsToPub)
}

type UrlQuery struct {
	url string
}

func (q *UrlQuery) Load(c *gin.Context) error {
	url, err := request.RequireQuery(c, "url")
	q.url = url
	return err
}

func ConnectWith(_ *request.NoParam, q *UrlQuery) (string, error) {
	connexionUrl := q.url + "/syn"
	status, body := ext.Call(
		ext.POST,
		connexionUrl,
		url.Values{},
		map[string]string{
			HeaderHost: request.Host,
		},
		"",
	)
	if status != 200 {
		return "", request.UnexpectedResponseFrom(connexionUrl, status, string(body))
	}
	connections[request.Host] = NAN
	return fmt.Sprintf("Demande de connexion envoyée à %s", q.url), nil
}

type HostHeader struct {
	host string
}

var _ request.Loadable = &HostHeader{}

func (h *HostHeader) Load(c *gin.Context) error {
	h.host = c.Request.Header.Get("Origin")
	log.Debug("got host \"%s\"", h.host)
	if h.host == "" {
		return request.ErrorResponse{
			Code: http.StatusBadRequest,
			Err:  "Veuillez renseigner un domaine valide dans le header \"Origin\"",
		}
	}
	return nil
}

func Syn(h *HostHeader, _ *request.NoParam) (string, error) {
	hostsToSub <- h.host
	return fmt.Sprintf("Syn request recieved, wait for syn-ack at \"%s\"", h.host), nil
}

func SendSynAck(hosts chan string) {
	log.Info("Starting SendSynAck worker")
	for host := range hosts {
		log.Debug("Starting new send syn-ack job for %s", host)
		connections[host] = SYN
		code, body := ext.Call(
			ext.POST, host+"/syn-ack", url.Values{}, map[string]string{
				HeaderHost: request.Host,
			},
			"Acknowledging Syn and waiting for acknowledgement of SynAck",
		)
		if code != 200 {
			log.Warn("Erreur pour envoyer syn-ack au host %s (%d - %s)", host, code, body)
		}
	}
}

func SynAck(h *HostHeader, _ *request.NoParam) (string, error) {
	hostsToPub <- h.host
	return fmt.Sprintf("Syn-Ack request recieved. Wait for ack at \"%s\".", h.host), nil
}

func SendAck(hosts chan string) {
	log.Info("Starting SendAck worker")
	for host := range hosts {
		log.Debug("Starting new send ack job at %s", host)
		connections[host] = ACK
		go KeepAlive(host)
		log.Debug("Acknowledging publisher %s", host)
		code, _ := ext.Call(
			ext.POST, host+"/ack", url.Values{}, map[string]string{
				HeaderHost: request.Host,
			},
			"Acknowledging successfull synack.",
		)
		if code != 200 {
			log.Debug("Erreur pour écouter le host %s (%d)", host, code)
		}
	}
}

func Ack(h *HostHeader, _ *request.NoParam) (string, error) {
	err := vetAuth(h.host, SYN)
	if err != nil {
		return "", err
	}
	log.Debug("Acknowledging %s", h.host)
	connections[h.host] = ACK
	go KeepAlive(h.host)
	return "Connexion successfull", nil
}

const DefaultWait = 5
const maxTries = 2

func KeepAlive(host string) {
	tries := 0
	for {
		status, ok := connections[host]
		if !ok {
			log.Debug("Connection with %s deleted", host)
		}
		if status != ACK {
			log.Debug("Host %s no longer acknowleged", host)
		}
		responseStatus, _ := ext.Call(ext.GET, host+"/health", url.Values{}, ext.Headers{}, "")
		if responseStatus == 200 {
			time.Sleep(DefaultWait * time.Second)
			tries = 0
		} else {
			log.Debug("Failed to keepalive %s (%d recieved)", host, status)
			tries++
			wait := 1 * time.Second
			for range tries {
				wait += wait
			}
			time.Sleep(wait)
			if tries > maxTries {
				log.Debug("Max number of retries expired, closing connection to %s", host)
				delete(connections, host)
				return
			} else {
				log.Debug("Retrying to connect with %s", host)
			}
		}
	}
}

func vetAuth(host string, expected connexionStatus) error {
	status, ok := connections[host]
	if !ok {
		log.Debug("Rejecting unregistered host \"%s\"", host)
		return request.ErrorResponse{
			Code: http.StatusUnauthorized,
			Err:  "Your host \"%s\" is not connected",
		}
	}
	if status != expected {
		log.Debug("Rejecting host %s, is %d instead of %d", host, status, expected)
		return request.ErrorResponse{
			Code: http.StatusUnauthorized,
			Err:  fmt.Sprintf("Your host \"%s\" has the status %d instead of %d", host, status, expected),
		}
	}
	return nil
}
