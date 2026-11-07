package synack

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Bruoche/messagerie-tr/internal/config"
	"github.com/Bruoche/messagerie-tr/pkg/ext"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/Bruoche/messagerie-tr/pkg/request"
	"github.com/gin-gonic/gin"
)

type connexionStatus int

const (
	ONG connexionStatus = iota
	SYN
	ACK
)

var statusLabelMapping = map[connexionStatus]string{
	ONG: "Contacted...",
	SYN: "Synchronizing...",
	ACK: "Connected",
}

func (s connexionStatus) ToString() string {
	label, ok := statusLabelMapping[s]
	if !ok {
		return fmt.Sprintf("Status %d", s)
	}
	return label
}

type Host string
type ID string

func (h Host) Request(path string) string {
	return string(h) + path
}

var hostsToSub = make(chan Host, 100)
var hostsToPub = make(chan Host, 100)
var connections = map[Host]connexionStatus{}
var registrations = map[ID]Host{}

// const KeyHeaderHost = "Origin"

func Serve() {
	go SendSynAck(hostsToSub)
	go SendAck(hostsToPub)
}

type HostQuery struct {
	url Host
}

func (q *HostQuery) Load(c *gin.Context) error {
	id, ok := c.GetQuery("id")
	if ok {
		url, ok := registrations[ID(id)]
		if ok {
			log.Debug("Resolved id \"%s\" as host %s", id, url)
			q.url = url
			return nil
		}
	}
	url, err := request.RequireQuery(c, "url")
	if err != nil {
		return err
	}
	log.Debug("Got host \"%s\" in query", url)
	q.url = Host(url)
	return nil
}

func ConnectWith(_ *request.NoParam, q *HostQuery) (string, error) {
	err := tryConnect(string(q.url) + "/broker/syn/" + config.Server.ID)
	if err != nil {
		// Possible que l'erreur vienne du fait qu'il s'agisse d'un node historique:
		rerr := tryConnect(string(q.url) + "/syn")
		if rerr != nil {
			return "", &request.ErrorResponse{
				Code: http.StatusInternalServerError,
				Err:  fmt.Sprintf("Both attempts to connect failed:\n- %s\n- %s", err, rerr),
			}
		}
		err := tryRegister(q.url)
		if err != nil {
			log.Debug("Impossible de s'enregistrer à %s, probablement legacy", q.url)
		}
	}
	connections[q.url] = ONG
	return fmt.Sprintf("Demande de connexion envoyée à %s", q.url), nil
}

func tryRegister(host Host) error {
	status, body := ext.Call(
		ext.POST,
		host.Request("/register/"+config.Server.ID),
		url.Values{},
		ext.Headers{},
		fmt.Sprintf("Send response to %s to acknowledge connection.", config.Server.Host+"/syn-ack"),
	)
	if status != http.StatusNoContent {
		return request.UnexpectedResponseFrom(string(host), status, string(body))
	}
	return nil
}

func tryConnect(endpoint string) error {
	log.Debug("Trying to connect at \"%s\"", endpoint)
	status, body := ext.Call(ext.POST, endpoint, url.Values{}, ext.Headers{}, "")
	if status != 202 {
		return request.UnexpectedResponseFrom(endpoint, status, string(body))
	}
	return nil
}

type HostHeader struct {
	host Host
}

var _ request.Loadable = &HostHeader{}

func (h *HostHeader) Load(c *gin.Context) error {
	h.host = Host(c.Request.Header.Get("Origin"))
	log.Debug("got host \"%s\"", h.host)
	if h.host == "" {
		return &request.ErrorResponse{
			Code: http.StatusBadRequest,
			Err:  "Veuillez renseigner un domaine valide dans le header \"Origin\"",
		}
	}
	return nil
}

type IdParam struct {
	id ID
}

func (p *IdParam) Load(c *gin.Context) error {
	id, err := request.RequireParam(c, "id")
	if err != nil {
		return err
	}
	p.id = ID(id)
	return nil
}

func SynRegistered(h *HostHeader, p *IdParam) error {
	register(h.host, p.id)
	return Syn(h, &request.NoParam{})
}

func Syn(h *HostHeader, _ *request.NoParam) error {
	hostsToSub <- h.host
	return nil
}

func SendSynAck(hosts chan Host) {
	log.Info("Starting SendSynAck worker")
	for host := range hosts {
		log.Debug("Starting new send syn-ack job for %s", host)
		connections[host] = SYN
		code, body := ext.Call(
			ext.POST, host.Request("/syn-ack"), url.Values{}, ext.Headers{},
			"Acknowledging Syn and waiting for acknowledgement of SynAck",
		)
		if code != 202 {
			log.Warn("Erreur pour envoyer syn-ack au host %s (%d - %s)", host, code, body)
		}
	}
}

func SynAck(h *HostHeader, _ *request.NoParam) error {
	err := vetAuth(h.host, ONG)
	if err != nil {
		return err
	}
	hostsToPub <- h.host
	return nil
}

func SendAck(hosts chan Host) {
	log.Info("Starting SendAck worker")
	for host := range hosts {
		log.Debug("Starting new send ack job at %s", host)
		connections[host] = ACK
		go KeepAlive(host)
		log.Debug("Acknowledging publisher %s", host)
		code, _ := ext.Call(ext.POST, host.Request("/ack"), url.Values{}, ext.Headers{},
			"Acknowledging successfull synack.",
		)
		if code != 204 {
			log.Debug("Erreur pour écouter le host %s (%d)", host, code)
		}
	}
}

func Ack(h *HostHeader, _ *request.NoParam) error {
	err := vetAuth(h.host, SYN)
	if err != nil {
		return err
	}
	log.Debug("Acknowledging %s", h.host)
	connections[h.host] = ACK
	go KeepAlive(h.host)
	return nil
}

const DefaultWait = 5
const maxTries = 7

func KeepAlive(host Host) {
	tries := 0
	for {
		status, ok := connections[host]
		if !ok {
			log.Debug("Connection with %s deleted", host)
			return
		}
		if status != ACK {
			log.Debug("Host %s no longer acknowleged", host)
			return
		}
		responseStatus, _ := ext.Call(ext.GET, host.Request("/health"), url.Values{}, ext.Headers{}, "")
		if responseStatus == http.StatusNoContent {
			time.Sleep(DefaultWait * time.Second)
			if tries > 0 {
				log.Debug("Keepalive with %s resolved.", host)
			}
			tries = 0
		} else {
			log.Debug("Failed to keepalive %s (%d recieved)", host, responseStatus)
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
				log.Debug("Retrying to connect with %s (retry %d/%d)", host, tries, maxTries)
			}
		}
	}
}

type IdBody struct {
	Host Host `json:"url"`
	Id   ID   `json:"id"`
}

var _ request.Loadable = &IdBody{}

func (q *IdBody) Load(c *gin.Context) error {
	return request.LoadJsonBody(c, q)
}

func RegisterBody(_ *request.NoParam, q *IdBody) error {
	register(q.Host, q.Id)
	return nil
}

func Register(h *HostHeader, q *IdParam) error {
	register(h.host, q.id)
	return nil
}

func register(host Host, id ID) {
	registrations[id] = host
}

func vetAuth(host Host, expected connexionStatus) error {
	status, ok := connections[host]
	if !ok {
		log.Debug("Rejecting unregistered host \"%s\"", host)
		return &request.ErrorResponse{
			Code: http.StatusUnauthorized,
			Err:  fmt.Sprintf("Host \"%s\" is not connected", host),
		}
	}
	if status != expected {
		log.Debug("Rejecting host %s, is %d instead of %d", host, status, expected)
		return &request.ErrorResponse{
			Code: http.StatusUnauthorized,
			Err:  fmt.Sprintf("Host \"%s\" has the status %d instead of %d", host, status, expected),
		}
	}
	return nil
}

func Connections(*request.NoParam, *request.NoParam) (string, error) {
	response := ""
	for host, status := range connections {
		idLabel := ""
		id, ok := getID(host)
		if ok {
			idLabel = fmt.Sprintf("(%s)", id)
		}
		response += fmt.Sprintf("%s%s: %s\n", host, idLabel, status.ToString())
	}
	if response == "" {
		return "No connexions made yet", nil
	}
	return response, nil
}

func Registrations(*request.NoParam, *request.NoParam) (string, error) {
	response := ""
	for id, host := range registrations {
		response += fmt.Sprintf("%s (%s)", id, host)
	}
	if response == "" {
		return "No host registered yet", nil
	}
	return response, nil
}

func getID(host Host) (ID, bool) {
	if host == Host(config.Server.Host) {
		id := config.Server.ID
		if id != "" {
			return ID(id), true
		}
	}
	for id, registered := range registrations {
		if host == registered {
			return id, true
		}
	}
	return ID(""), false
}

func Disconnect(_ *request.NoParam, q *HostQuery) (string, error) {
	err := vetAuth(q.url, ACK)
	if err != nil {
		return "", err
	}
	code, body := ext.Call(ext.POST, q.url.Request("/forget"), url.Values{}, ext.Headers{}, "")
	if code != http.StatusNoContent {
		return "", request.UnexpectedResponseFrom(string(q.url), code, string(body))
	}
	delete(connections, q.url)
	return fmt.Sprintf("Connexion avec %s terminée", q.url), nil
}

func Forget(h *HostHeader, expected *request.NoParam) error {
	err := vetAuth(h.host, ACK)
	if err != nil {
		return err
	}
	delete(connections, h.host)
	return nil
}
