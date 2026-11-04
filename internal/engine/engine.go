package engine

import (
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/Bruoche/messagerie-tr/pkg/request"
	"github.com/gin-gonic/gin"
)

var engine *gin.Engine
var Host string

type Route struct {
	endpoint string
	handler  gin.HandlerFunc
	method   request.Method
}

func MakeRoute(method request.Method, endpoint string, handler gin.HandlerFunc) Route {
	return Route{
		endpoint: endpoint,
		handler:  handler,
		method:   method,
	}
}

func mapGet(route Route) {
	engine.GET(route.endpoint, route.handler)
}
func mapPost(route Route) {
	engine.POST(route.endpoint, route.handler)
}

var httpMethodMapRouteHandler = map[request.Method]func(Route){
	request.GET:  mapGet,
	request.POST: mapPost,
}

func Boot(host string, routes []Route) {
	engine = gin.New()
	engine.SetTrustedProxies(nil)
	for _, route := range routes {
		handleRoute, ok := httpMethodMapRouteHandler[route.method]
		if ok {
			handleRoute(route)
		} else {
			log.Error("La méthode \"%s\" (de la route %s) n'est pas gérée", route.method, route.endpoint)
		}
	}
	Host = "http://" + host
	engine.Run(host)
}
