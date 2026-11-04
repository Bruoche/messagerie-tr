package synack

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Bruoche/messagerie-tr/internal/engine"
	"github.com/Bruoche/messagerie-tr/internal/logserv"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/Bruoche/messagerie-tr/pkg/request"
	"github.com/gin-gonic/gin"
)

type subStatus int

const (
	NAN subStatus = iota
	SYN
	ACK
)

var hostsToSub = make(chan string, 100)
var hostsToPub = make(chan string, 100)
var connections = map[string]subStatus{}

const HeaderHost = "Origin"

func Start() {
	go SendSynAck(hostsToSub)
	go SendAck(hostsToPub)
}

func ConnectWith(context *gin.Context) {
	logserv.Called("/sub")
	targetURL, ok := context.GetQuery("url")
	if !ok {
		context.String(http.StatusBadRequest, "Veuillez renseigner le domaine où s'abonner (champs \"url\")")
		return
	}
	status, raw := request.Call(
		request.POST,
		targetURL+"/syn",
		url.Values{},
		map[string]string{
			HeaderHost: engine.Host,
		},
		"",
	)
	body := string(raw)
	if status == 200 {
		body = fmt.Sprintf("Connexion réussi avec %s", targetURL)
		connections[engine.Host] = NAN
	}
	context.String(status, body)
}

func Syn(context *gin.Context) {
	logserv.Called("/syn")
	sub, ok := getHostOrHandle(context)
	if !ok {
		return
	}
	hostsToSub <- sub
	context.String(http.StatusOK, fmt.Sprintf("Syn request recieved, wait for syn-ack at \"%s\"", sub))
}

func SendSynAck(hosts chan string) {
	log.Info("Starting SendSynAck worker")
	for host := range hosts {
		log.Debug("Starting new send syn-ack job at %s", host)
		connections[host] = SYN
		log.Debug("Syncing subscriber %s", host)
		code, _ := request.Call(
			request.POST, host+"/syn-ack", url.Values{}, map[string]string{
				HeaderHost: engine.Host,
			},
			"Acknowledging Syn and waiting for acknowledgement of SynAck",
		)
		if code != 200 {
			log.Debug("Erreur pour se connecter au host %s (%d)", host, code)
		}
	}
}

func SynAck(context *gin.Context) {
	logserv.Called("/syn-ack")
	host, ok := getHostOrHandle(context)
	if !ok {
		return
	}
	hostsToPub <- host
	context.String(http.StatusOK, fmt.Sprintf("Syn-Ack request recieved. Wait for ack at \"%s\".", host))
}

func SendAck(hosts chan string) {
	log.Info("Starting SendAck worker")
	for host := range hosts {
		log.Debug("Starting new send ack job at %s", host)
		connections[host] = ACK
		go KeepAlive(host)
		log.Debug("Acknowledging publisher %s", host)
		code, _ := request.Call(
			request.POST, host+"/ack", url.Values{}, map[string]string{
				HeaderHost: engine.Host,
			},
			"Acknowledging successfull synack.",
		)
		if code != 200 {
			log.Debug("Erreur pour écouter le host %s (%d)", host, code)
		}
	}
}

func Ack(context *gin.Context) {
	logserv.Called("/ack")
	host, ok := getHostOrHandle(context)
	if !ok {
		return
	}
	status, ok := connections[host]
	if !ok {
		log.Debug("Rejecting from ack unregistered host \"%s\"", host)
		reject(context, host)
	}
	if status != SYN {
		log.Debug("Rejecting from ack host \"%s\" of status %d instead of %d", host, status, SYN)
		reject(context, host)
	}
	log.Debug("Acknowledging %s", host)
	connections[host] = ACK
	KeepAlive(host)
}

const DefaultWait = 5
const maxTries = 10

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
		responseStatus, _ := request.Call(request.GET, host+"/health", url.Values{}, request.Headers{}, "")
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

func AuthOrReject(context *gin.Context) (host string, ok bool) {
	host, ok = getHostOrHandle(context)
	if !ok {
		return "", false
	}
	status, ok := connections[host]
	if !ok {
		log.Debug("Rejecting unregistered host \"%s\"", host)
		reject(context, host)
		return
	}
	if status != ACK {
		log.Debug("Rejecting host %s, is %d instead of %d", host, status, ACK)
		reject(context, host)
		return "", false
	}
	log.Debug("Authentified host %s", host)
	return host, true
}

func getHostOrHandle(context *gin.Context) (host string, ok bool) {
	sub := context.Request.Header.Get("Origin")
	log.Debug("got host \"%s\"", sub)
	if sub == "" {
		context.String(
			http.StatusBadRequest,
			"Veuillez renseigner un domaine valide dans le header \"Origin\"",
		)
		return "", false
	}
	return sub, true
}

func reject(context *gin.Context, host string) {
	context.String(http.StatusUnauthorized, "Your host \"%s\" has not been acknowledged", host)
}
