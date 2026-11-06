package main

import (
	"flag"
	"strconv"

	"github.com/Bruoche/messagerie-tr/internal/msg"
	"github.com/Bruoche/messagerie-tr/internal/synack"
	"github.com/Bruoche/messagerie-tr/pkg/ext"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/Bruoche/messagerie-tr/pkg/request"
)

func main() {
	log.Info("Starting application...")
	synack.Start()
	var port int
	flag.IntVar(&port, "p", 8085, "Port used to start the server")
	var host string
	flag.StringVar(&host, "h", "", "Host sent as Origin")
	var id string
	flag.StringVar(&id, "id", "", "ID to be registered as on the network.")
	flag.Parse()
	if host == "" {
		host = "http://localhost:" + strconv.Itoa(port)
	}
	request.Init(
		host, port, id,
		request.MakePath(ext.GET, "/", request.String(request.Index)),
		request.MakePath(ext.GET, "/health", request.NoContent(healthCheck)),
		request.MakePath(ext.GET, "/message", request.String(msg.Get)),
		request.MakePath(ext.GET, "/message/from", request.String(msg.Call)),
		request.MakePath(ext.GET, "/connect", request.String(synack.ConnectWith)),
		request.MakePath(ext.GET, "/connections", request.String(synack.Connections)),
		request.MakePath(ext.POST, "/syn", request.Accept(synack.Syn)),
		request.MakePath(ext.POST, "/syn-ack", request.Accept(synack.SynAck)),
		request.MakePath(ext.POST, "/ack", request.NoContent(synack.Ack)),
		request.MakePath(ext.GET, "/publish", request.String(synack.Send)),
		request.MakePath(ext.POST, "/ingress", request.Created(synack.Recieve)),
		request.MakePath(ext.GET, "/messages", request.String(synack.Messages)),
		request.MakePath(ext.POST, "/forget", request.NoContent(synack.Forget)),
	)
}

func healthCheck(*request.NoParam, *request.NoParam) error {
	return nil
}
