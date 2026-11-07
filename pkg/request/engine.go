package request

import (
	"fmt"

	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/gin-gonic/gin"
)

var engine *gin.Engine
var Paths []Pathable

func Serve(domain string, showDebug bool, routes ...Pathable) {
	if !showDebug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine = gin.New()
	engine.SetTrustedProxies(nil)
	for _, route := range routes {
		handlePath, ok := httpMethodMapPathHandler[route.getMethod()]
		if ok {
			handlePath(route)
			Paths = append(Paths, route)
		} else {
			log.Error("La méthode \"%s\" (du chemin %s) n'est pas gérée", route.getMethod(), route.getEndpoint())
		}
	}
	log.Info("Starting engine on %s", domain)
	engine.Run(domain)
}

func Index(*NoParam, *NoParam) (string, error) {
	routeList := ""
	for _, route := range Paths {
		routeList += fmt.Sprintf("%s - %s\n", route.getMethod(), route.getEndpoint())
	}
	return routeList, nil
}
