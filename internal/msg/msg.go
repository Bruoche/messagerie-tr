package msg

import (
	"net/http"
	"net/url"

	"github.com/Bruoche/messagerie-tr/internal/logserv"
	"github.com/Bruoche/messagerie-tr/pkg/request"
	"github.com/gin-gonic/gin"
)

func Call(context *gin.Context) {
	logserv.Called("/call")
	targetURL, ok := context.GetQuery("url")
	if !ok {
		context.String(http.StatusBadRequest, "Veuillez renseigner l'url appelée")
		return
	}
	status, body := request.Call(request.GET, targetURL+"/message", url.Values{}, map[string]string{}, "")
	context.String(status, string(body))
}

func Get(context *gin.Context) {
	logserv.Called("/message")
	context.String(http.StatusOK, "hello")
}
