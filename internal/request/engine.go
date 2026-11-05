package request

import (
	"fmt"

	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/gin-gonic/gin"
)

var engine *gin.Engine
var Host string
var Paths []Pathable

func Init(host string, routes ...Pathable) {
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
	Host = "http://" + host
	engine.Run(host)
}

func Index(*NoParam, *NoParam) (string, error) {
	routeList := ""
	for _, route := range Paths {
		routeList += fmt.Sprintf("%s - %s\n", route.getMethod(), route.getEndpoint())
	}
	return routeList, nil
}
