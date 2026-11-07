package request

import (
	"fmt"
	"net/http"

	"github.com/Bruoche/messagerie-tr/pkg/log"
)

type ErrorResponse struct {
	Code int
	Err  string
}

func (e ErrorResponse) Error() string {
	return e.Err
}

func UnexpectedResponseFrom(url string, code int, response string) *ErrorResponse {
	log.Debug("Unexpected response obtained from %s: %d - \"%s\")", url, code, response)
	return &ErrorResponse{
		Code: http.StatusInternalServerError,
		Err:  fmt.Sprintf("Got unexpected status %d from %s", code, url),
	}
}
