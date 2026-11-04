package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/Bruoche/messagerie-tr/internal/engine"
	"github.com/Bruoche/messagerie-tr/internal/logserv"
	"github.com/Bruoche/messagerie-tr/internal/msg"
	"github.com/Bruoche/messagerie-tr/internal/synack"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/Bruoche/messagerie-tr/pkg/request"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Info("Starting application...")
	synack.Start()
	var port int
	flag.IntVar(&port, "p", 8085, "Port used to start the server")
	flag.Parse()
	engine.Boot("localhost:"+strconv.Itoa(port), []engine.Route{
		engine.MakeRoute(request.GET, "/health", healthCheck),
		engine.MakeRoute(request.GET, "/message", msg.Get),
		engine.MakeRoute(request.GET, "/message/from", msg.Call),
		engine.MakeRoute(request.GET, "/sub", synack.ConnectWith),
		engine.MakeRoute(request.POST, "/syn", synack.Syn),
		engine.MakeRoute(request.POST, "/syn-ack", synack.SynAck),
		engine.MakeRoute(request.POST, "/ack", synack.Ack),
		engine.MakeRoute(request.GET, "/publish", synack.Send),
		engine.MakeRoute(request.POST, "/ingress", synack.Recieve),
	})
}

func healthCheck(context *gin.Context) {
	logserv.Called("health check")
	context.String(http.StatusOK, "healthy")
}
