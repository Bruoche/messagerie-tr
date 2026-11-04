package logserv

import (
	"fmt"
	"time"
)

func Called(request string) {
	fmt.Printf(
		"[SERVER] %s - Request %s called\n",
		time.Now().Format("02/01/2006 - 15:04:05"),
		request,
	)
}
