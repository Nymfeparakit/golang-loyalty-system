package domain

import "net/http"

// ResponseWithReadBody http ответ, который в качестве отдельного поля содержит байты уже прочитанного тела
type ResponseWithReadBody struct {
	Response *http.Response
	ReadBody []byte
}
