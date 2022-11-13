package domain

import "net/http"

type ResponseWithReadBody struct {
	Response *http.Response
	ReadBody []byte
}
