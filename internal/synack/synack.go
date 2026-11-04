package synack

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Bruoche/messagerie-tr/internal/engine"
	"github.com/Bruoche/messagerie-tr/internal/logserv"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/Bruoche/messagerie-tr/pkg/request"
	"github.com/gin-gonic/gin"
)

type subStatus int
type pubStatus int

const (
	SYN subStatus = iota
	ACK
	WAIT pubStatus = iota
	SA
)

var hostsToSub = make(chan string, 100)
var subscribers = map[string]subStatus{}
var hostsToPub = make(chan string, 100)
var listens = map[string]pubStatus{}

const HeaderHost = "Origin"

func Start() {
	go SendSynAck(hostsToSub)
	go SendAck(hostsToPub)
}

func SubTo(context *gin.Context) {
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
		listens[engine.Host] = WAIT
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
		subscribers[host] = SYN
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
	pub, ok := getHostOrHandle(context)
	if !ok {
		return
	}
	hostsToPub <- pub
	context.String(http.StatusOK, fmt.Sprintf("Syn-Ack request recieved. Wait for ack at \"%s\".", pub))
}

func SendAck(hosts chan string) {
	log.Info("Starting SendAck worker")
	for host := range hosts {
		log.Debug("Starting new send ack job at %s", host)
		listens[host] = SA
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
	sub, ok := getHostOrHandle(context)
	if !ok {
		return
	}
	log.Debug("Acknowledging %s", sub)
	subscribers[sub] = ACK
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
