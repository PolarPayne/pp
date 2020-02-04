package pp

import "net/http"

type Auth interface {
	HandleAuth(http.ResponseWriter, *http.Request, Storage, bool) error
}
