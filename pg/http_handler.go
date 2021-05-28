package pg

import (
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// HTTPHandler is DB queries HTTP handler.
type HTTPHandler struct {
	DB *sqlx.DB
}

type httpRequest struct {
	QueryName string        `json:"query_name"`
	Args      []interface{} `json:"args"`
}

// Handle handles HTTP requests for the select queries.
func (h HTTPHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req httpRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest)
		return
	}

	res, err := Select(r.Context(), h.DB, req.QueryName, req.Args)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}
