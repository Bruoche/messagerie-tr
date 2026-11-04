package request

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/Bruoche/messagerie-tr/pkg/log"
)

type Method string

const (
	GET  Method = "GET"
	POST Method = "POST"
)

func Call(method Method, baseUrl string, params url.Values, headers map[string]string, body any) (int, []byte) {
	response := RawCall(string(method), baseUrl, params, headers, body)
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error("Une erreur est survenue lors de la lecture du body de la requête")
	}
	defer response.Body.Close()
	return response.StatusCode, responseBody
}

func RawCall(method string, baseUrl string, params url.Values, headers map[string]string, body any) *http.Response {
	url := baseUrl + "?" + params.Encode()
	bodyBuffer := new(bytes.Buffer)
	if err := json.NewEncoder(bodyBuffer).Encode(body); err != nil {
		log.Error("Erreur d'encodage du body à l'url \"%s\" :\n  >> %s", url, err)
	}
	request, err := http.NewRequest(method, url, bodyBuffer)
	if err != nil {
		log.Error("Une erreur est survenue pour la création de la requête \"%s\" :\n  >> %s", url, err)
	}
	for header, value := range headers {
		request.Header.Add(header, value)
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Error("Une erreur est survenue lors de l'envois de la requête \"%s\" :\n  >> %s", url, err)
	}
	log.Debug("(%d) Requête \"%s\" traitée, réponse obtenue:\n%v\nBody:%s", response.StatusCode, url, response, request.Body)
	return response
}
