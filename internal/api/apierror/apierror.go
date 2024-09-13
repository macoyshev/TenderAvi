package apierror

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Reason string `json:"reason"`
}

func HandleError(w http.ResponseWriter, r *http.Request, err error, status int) {
	errRes := ErrorResponse{Reason: err.Error()}
	res, _ := json.Marshal(errRes)
	w.WriteHeader(status)
	w.Write(res)
}
