package main

import (
	"flag"

	"github.com/Bruoche/messagerie-tr/internal/config"
	"github.com/Bruoche/messagerie-tr/internal/msg"
	"github.com/Bruoche/messagerie-tr/internal/synack"
	"github.com/Bruoche/messagerie-tr/pkg/ext"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/Bruoche/messagerie-tr/pkg/request"
)

func main() {
	args := getArgs()
	log.Info("Starting application...")
	setConfig(args)
	synack.Serve()
	request.Serve(
		config.Server.Domain, config.Debug,
		request.MakePath(ext.GET, "/", request.String(request.Index)),
		request.MakePath(ext.GET, "/health", request.NoContent(healthCheck)),
		request.MakePath(ext.GET, "/message", request.String(msg.Get)),
		request.MakePath(ext.GET, "/message/from", request.String(msg.Call)),
		request.MakePath(ext.GET, "/connect", request.String(synack.ConnectWith)),
		request.MakePath(ext.GET, "/disconnect", request.String(synack.Disconnect)),
		request.MakePath(ext.GET, "/connections", request.String(synack.Connections)),
		request.MakePath(ext.GET, "/registrations", request.String(synack.Registrations)),
		request.MakePath(ext.GET, "/publish", request.String(synack.Send)),
		request.MakePath(ext.GET, "/messages", request.String(synack.Messages)),
		request.MakePath(ext.POST, "/register", request.NoContent(synack.RegisterBody)),
		request.MakePath(ext.POST, "/register/:id", request.NoContent(synack.Register)),
		request.MakePath(ext.POST, "/broker/syn/:id", request.Accept(synack.SynRegistered)),
		request.MakePath(ext.POST, "/syn", request.Accept(synack.Syn)),
		request.MakePath(ext.POST, "/syn-ack", request.Accept(synack.SynAck)),
		request.MakePath(ext.POST, "/ack", request.NoContent(synack.Ack)),
		request.MakePath(ext.POST, "/ingress", request.Created(synack.Recieve)),
		request.MakePath(ext.POST, "/forget", request.NoContent(synack.Forget)),
	)
}

func setConfig(args Args) {
	config.Set(args.host, args.ID, args.port, args.debug)
	log.ShowDebug = config.Debug
	ext.Host = config.Server.Host
}

type Args struct {
	port  int
	host  string
	ID    string
	debug bool
}

func getArgs() Args {
	var f Args
	flag.IntVar(&f.port, "p", 8085, "Port used to start the server")
	flag.StringVar(&f.host, "as", "", "Host sent as Origin")
	flag.StringVar(&f.ID, "id", "", "ID to be registered as on the network.")
	flag.BoolVar(&f.debug, "debug", false, "Shows debug logs if set.")
	flag.Parse()
	return f
}

func healthCheck(*request.NoParam, *request.NoParam) error {
	return nil
}
