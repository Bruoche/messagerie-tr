package synack

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Bruoche/messagerie-tr/pkg/ext"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/Bruoche/messagerie-tr/pkg/request"
	"github.com/gin-gonic/gin"
)

type MessageQuery struct {
	message string
	target  string
}

func (q *MessageQuery) Load(c *gin.Context) error {
	message, err := request.RequireQuery(c, "message")
	target, err := request.RequireQuery(c, "target")
	q.message = message
	q.target = target
	return err
}

type Metadata struct {
	Source    string `json:"source"`
	Target    string `json:"target"`
	Timestamp string `json:"timestamp"`
}

type MessageResponse struct {
	Metadata Metadata `json:"metadata"`
	Action   string   `json:"action"`
	Payload  string   `json:"payload"`
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
	err := vetAuth(q.target, ACK)
	if err != nil {
		return "", err
	}
	msgBody := MessageResponse{
		Metadata: Metadata{
			Source:    request.Host,
			Target:    q.target,
			Timestamp: time.Now().Format("2006-01-02T15:04:05.000Z"),
		},
		Action:  "MSG",
		Payload: q.message,
	}
	status, response := ext.Call(
		ext.POST, q.target+"/ingress", url.Values{}, map[string]string{
			HeaderHost: request.Host,
		},
		msgBody,
	)
	log.Debug("Sent \"%+s\" to %s", q.message, q.target)
	if status != 202 {
		return "", fmt.Errorf("Couldn't send message to %s: Got (%d) %s", q.target, status, response)
	}
	messages = append(messages, RecievedMessage{msgBody, request.Host})
	return "Message sent to %s.", nil
}

func Recieve(h *HostHeader, r *MessageResponse) error {
	err := vetAuth(h.host, ACK)
	if err != nil {
		return err
	}
	log.Debug("Got message \"%s\"", r.Payload)
	messages = append(messages, RecievedMessage{*r, h.host})
	return nil
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
