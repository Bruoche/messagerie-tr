package synack

import (
	"fmt"
	"time"

	"github.com/Bruoche/messagerie-tr/internal/config"
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

type MessageDto struct {
	Metadata Metadata `json:"metadata"`
	Action   string   `json:"action"`
	Payload  string   `json:"payload"`
}

var _ request.Loadable = &MessageDto{}

func (r *MessageDto) Load(c *gin.Context) error {
	return request.LoadJsonBody(c, r)
}

var messages = []MessageDto{}

func Send(_ *request.NoParam, q *MessageQuery) (string, error) {
	host, ok := registrations[ID(q.target)]
	if !ok {
		host = Host(q.target)
	}
	err := vetAuth(host, ACK)
	if err != nil {
		return "", err
	}
	msgBody := MessageDto{
		Metadata: Metadata{
			Source:    config.Server.ID,
			Target:    q.target,
			Timestamp: time.Now().Format("2006-01-02T15:04:05.000Z"),
		},
		Action:  "MSG",
		Payload: q.message,
	}
	status, response := ext.SimpleCall(
		ext.POST, host.Request("/ingress"), msgBody,
	)
	log.Debug("Sent \"%+s\" to %s", q.message, q.target)
	if status != 202 {
		return "", fmt.Errorf("Couldn't send message to %s: Got (%d) %s", host, status, response)
	}
	messages = append(messages, msgBody)
	return fmt.Sprintf("Message sent to %s.", q.target), nil
}

func Recieve(h *HostHeader, r *MessageDto) error {
	err := vetAuth(h.host, ACK)
	if err != nil {
		return err
	}
	log.Debug("Got message \"%s\"", r.Payload)
	messages = append(messages, *r)
	return nil
}

func Messages(*request.NoParam, *request.NoParam) (string, error) {
	board := ""
	for _, message := range messages {
		source := message.Metadata.Source
		id, ok := getID(Host(message.Metadata.Source))
		if ok {
			source = string(id)
		}
		board += fmt.Sprintf("%s [%s]: %s\n", source, message.Action, message.Payload)
	}
	if board == "" {
		return "No messages sent or recieved so far", nil
	}
	return board, nil
}
