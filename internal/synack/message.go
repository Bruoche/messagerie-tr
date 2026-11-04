package synack

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Bruoche/messagerie-tr/internal/engine"
	"github.com/Bruoche/messagerie-tr/internal/logserv"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/Bruoche/messagerie-tr/pkg/request"
	"github.com/gin-gonic/gin"
)

type Message struct {
	Action  string
	Payload string
}

var messages = []Message{}

func Send(context *gin.Context) {
	logserv.Called("/send")
	message, ok := context.GetQuery("message")
	if !ok {
		context.String(http.StatusBadRequest, "Veuillez renseigner le paramètre \"message\" dans la requête")
	}
	msgBody := Message{"MSG", message}
	errs := []error{}
	recieved := false
	for host, ack := range connections {
		if ack == ACK {
			status, response := request.Call(
				request.POST, host+"/ingress", url.Values{}, map[string]string{
					HeaderHost: engine.Host,
				},
				msgBody,
			)
			log.Debug("Sent \"%+s\" to %s", message, host)
			if status != 200 {
				errs = append(errs, fmt.Errorf("%s: (%d) %s", host, status, response))
			} else {
				recieved = true
			}
		}
	}
	if recieved {
		messages = append(messages, msgBody)
	}
	body := "Message sent to all connections."
	err := errors.Join(errs...)
	if err != nil {
		body = fmt.Sprintf("Some hosts failed:\n%s", err)
	}
	context.String(http.StatusOK, body)
}

func Recieve(context *gin.Context) {
	logserv.Called("/ingress")
	_, ok := AuthOrReject(context)
	if !ok {
		return
	}
	var message Message
	err := context.ShouldBindJSON(&message)
	if err != nil {
		log.Error("%s", err)
		context.String(
			http.StatusBadRequest,
			fmt.Sprintf("Impossible de déserialiser le corps de la requête: %s", err),
		)
		return
	}
	log.Debug("Got message \"%s\"", message.Payload)
	context.String(http.StatusOK, "Message recieved.")
}
