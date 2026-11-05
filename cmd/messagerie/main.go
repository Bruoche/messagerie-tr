package main

import (
	"flag"
	"strconv"

	"github.com/Bruoche/messagerie-tr/internal/msg"
	request "github.com/Bruoche/messagerie-tr/internal/request"
	"github.com/Bruoche/messagerie-tr/internal/synack"
	"github.com/Bruoche/messagerie-tr/pkg/ext"
	"github.com/Bruoche/messagerie-tr/pkg/log"
)

func main() {
	log.Info("Starting application...")
	synack.Start()
	var port int
	flag.IntVar(&port, "p", 8085, "Port used to start the server")
	flag.Parse()
	request.Init(
		"localhost:"+strconv.Itoa(port),
		request.MakePath(ext.GET, "/", request.String(request.Index)),
		request.MakePath(ext.GET, "/health", request.String(healthCheck)),
		request.MakePath(ext.GET, "/message", request.String(msg.Get)),
		request.MakePath(ext.GET, "/message/from", request.String(msg.Call)),
		request.MakePath(ext.GET, "/connect", request.String(synack.ConnectWith)),
		request.MakePath(ext.POST, "/syn", request.String(synack.Syn)),
		request.MakePath(ext.POST, "/syn-ack", request.String(synack.SynAck)),
		request.MakePath(ext.POST, "/ack", request.String(synack.Ack)),
		request.MakePath(ext.GET, "/publish", request.String(synack.Send)),
		request.MakePath(ext.POST, "/ingress", request.String(synack.Recieve)),
		request.MakePath(ext.GET, "/messages", request.String(synack.Messages)),
	)
}

func healthCheck(*request.NoParam, *request.NoParam) (string, error) {
	return "healthy", nil
}
