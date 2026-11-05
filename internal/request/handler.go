package request

import (
	"fmt"
	"net/http"

	"github.com/Bruoche/messagerie-tr/internal/logserv"
	"github.com/Bruoche/messagerie-tr/pkg/ext"
	"github.com/Bruoche/messagerie-tr/pkg/log"
	"github.com/gin-gonic/gin"
)

type Pathable interface {
	getHandler() gin.HandlerFunc
	getEndpoint() string
	getMethod() ext.Method
}
type Path[H, Q any, PH RequestParam[H], PQ RequestParam[Q]] struct {
	endpoint string
	handler  Handler[H, Q, PH, PQ]
	method   ext.Method
}

var _ Pathable = Path[NoParam, NoParam, *NoParam, *NoParam]{}

func (p Path[H, Q, PH, PQ]) getHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logserv.Called(p.endpoint)
		headers, err := p.handler.getHeader(c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}
		query, err := p.handler.getQuery(c)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}
		err = p.handler.respond(c, headers, query)
		if err != nil {
			rerr, ok := err.(*ErrorResponse)
			if ok {
				c.String(rerr.Code, rerr.Error())
				return
			}
			log.Error("Erreur lors de la requête %s: %s", p.endpoint, err)
			c.String(http.StatusInternalServerError, "Une erreur est survenue.")
		}
	}
}
func (p Path[H, Q, PH, PQ]) getEndpoint() string {
	return p.endpoint
}
func (p Path[H, Q, PH, PQ]) getMethod() ext.Method {
	return p.method
}

type Handler[H, Q any, PH RequestParam[H], PQ RequestParam[Q]] interface {
	getHeader(*gin.Context) (PH, error)
	getQuery(*gin.Context) (PQ, error)
	respond(*gin.Context, PH, PQ) error
}
type StringRequest[H, Q any, PH RequestParam[H], PQ RequestParam[Q]] struct {
	getterH func(*gin.Context) (PH, error)
	getterQ func(*gin.Context) (PQ, error)
	handler func(PH, PQ) (string, error)
}

var _ Handler[NoParam, NoParam, *NoParam, *NoParam] = StringRequest[NoParam, NoParam, *NoParam, *NoParam]{}

func String[H, Q any, PH RequestParam[H], PQ RequestParam[Q]](handler func(PH, PQ) (string, error)) StringRequest[H, Q, PH, PQ] {
	return StringRequest[H, Q, PH, PQ]{
		makeGetter[H, PH],
		makeGetter[Q, PQ],
		handler,
	}
}
func (r StringRequest[H, Q, PH, PQ]) getHeader(c *gin.Context) (PH, error) {
	return r.getterH(c)
}
func (r StringRequest[H, Q, PH, PQ]) getQuery(c *gin.Context) (PQ, error) {
	return r.getterQ(c)
}
func (r StringRequest[H, Q, PH, PQ]) respond(c *gin.Context, header PH, query PQ) error {
	response, err := r.handler(header, query)
	if err != nil {
		return err
	}
	c.String(http.StatusOK, response)
	return nil
}

type Loadable interface {
	Load(*gin.Context) error
}
type RequestParam[T any] interface {
	*T
	Loadable
}

func makeGetter[T any, P RequestParam[T]](c *gin.Context) (P, error) {
	var new T
	ptr := P(&new)
	err := ptr.Load(c)
	return ptr, err
}

func RequireQuery(c *gin.Context, name string) (string, error) {
	str, ok := c.GetQuery(name)
	if !ok {
		return "", ErrorResponse{
			Code: http.StatusBadRequest,
			Err:  fmt.Sprintf("Veuillez renseigner le paramètre requis \"%s\"", name),
		}
	}
	return str, nil
}

// No Header/Query parameter
type NoParam struct{}

var _ Loadable = NoParam{}

func (n NoParam) Load(*gin.Context) error {
	return nil
}

func MakePath[H, Q any, PH RequestParam[H], PQ RequestParam[Q]](method ext.Method, endpoint string, handler Handler[H, Q, PH, PQ]) Pathable {
	return Path[H, Q, PH, PQ]{
		endpoint: endpoint,
		handler:  handler,
		method:   method,
	}
}

func mapGet(route Pathable) {
	engine.GET(route.getEndpoint(), route.getHandler())
}
func mapPost(route Pathable) {
	engine.POST(route.getEndpoint(), route.getHandler())
}

var httpMethodMapPathHandler = map[ext.Method]func(Pathable){
	ext.GET:  mapGet,
	ext.POST: mapPost,
}
