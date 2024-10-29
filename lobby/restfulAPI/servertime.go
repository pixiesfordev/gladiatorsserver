package restfulAPI

import (
	"encoding/json"
	"net/http"
	"time"
)

// [GET] /game/servertime
func ServerTime(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"data": time.Now(),
	}
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
