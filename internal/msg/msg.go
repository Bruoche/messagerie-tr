package msg

import (
	"net/url"

	"github.com/Bruoche/messagerie-tr/pkg/ext"
	"github.com/Bruoche/messagerie-tr/pkg/request"
	"github.com/gin-gonic/gin"
)

type UrlQuery struct {
	url string
}

var _ request.Loadable = &UrlQuery{}

func (q *UrlQuery) Load(c *gin.Context) error {
	url, err := request.RequireQuery(c, "url")
	q.url = url
	return err
}

func Call(_ *request.NoParam, q *UrlQuery) (string, error) {
	targetUrl := q.url + "/message"
	status, body := ext.Call(ext.GET, targetUrl, url.Values{}, map[string]string{}, "")
	if status != 200 {
		return "", request.UnexpectedResponseFrom(targetUrl, status, string(body))
	}
	return string(body), nil
}

func Get(*request.NoParam, *request.NoParam) (string, error) {
	return "hello", nil
}
