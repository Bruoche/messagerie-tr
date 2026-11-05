package synack

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Bruoche/messagerie-tr/internal/request"
	"github.com/Bruoche/messagerie-tr/pkg/ext"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/gin-gonic/gin"
)

type MessageQuery struct {
	message string
}

func (q *MessageQuery) Load(c *gin.Context) error {
	message, err := request.RequireQuery(c, "message")
	q.message = message
	return err
}

type MessageResponse struct {
	Action  string
	Payload string
}

func (r *MessageResponse) Load(c *gin.Context) error {
	err := c.ShouldBindJSON(r)
	if err != nil {
		return request.ErrorResponse{
			Code: http.StatusBadRequest,
			Err:  fmt.Sprintf("Impossible de déserialiser le corps de la requête: %s", err),
		}
	}
	return nil
}

type RecievedMessage struct {
	MessageResponse
	host string
}

var messages = []RecievedMessage{}

func Send(_ *request.NoParam, q *MessageQuery) (string, error) {
	msgBody := MessageResponse{"MSG", q.message}
	errs := []error{}
	recieved := false
	for host, ack := range connections {
		if ack == ACK {
			status, response := ext.Call(
				ext.POST, host+"/ingress", url.Values{}, map[string]string{
					HeaderHost: request.Host,
				},
				msgBody,
			)
			log.Debug("Sent \"%+s\" to %s", q.message, host)
			if status != 200 {
				errs = append(errs, fmt.Errorf("%s: (%d) %s", host, status, response))
			} else {
				recieved = true
			}
		}
	}
	if recieved {
		messages = append(messages, RecievedMessage{msgBody, request.Host})
	}
	err := errors.Join(errs...)
	if err != nil {
		return fmt.Sprintf("Some hosts failed:\n%s", err), nil
	}
	return "Message sent to all connections.", nil
}

func Recieve(h *HostHeader, r *MessageResponse) (string, error) {
	err := vetAuth(h.host, ACK)
	if err != nil {
		return "", err
	}
	log.Debug("Got message \"%s\"", r.Payload)
	messages = append(messages, RecievedMessage{*r, h.host})
	return "Message recieved.", nil
}

func Messages(*request.NoParam, *request.NoParam) (string, error) {
	board := ""
	for _, message := range messages {
		board += fmt.Sprintf("%s [%s]: %s\n", message.host, message.Action, message.Payload)
	}
	if board == "" {
		return "No messages sent or recieved so far", nil
	}
	return board, nil
}
