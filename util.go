package codenames

import (
	"encoding/json"
	"net/http"
)

func writeJSON(rw http.ResponseWriter, resp interface{}) {
	j, err := json.Marshal(resp)
	if err != nil {
		http.Error(rw, "unable to marshal response", 500)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(j)
}
