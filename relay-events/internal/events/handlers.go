package events

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Server wires HTTP handlers to a Store.
type Server struct {
	Store *Store
}

// NewServer returns a Server backed by a fresh in-memory Store.
func NewServer() *Server {
	return &Server{Store: NewStore()}
}

// HandleEvents implements POST /events.
//
// The whole batch always gets a 200: the batch was received and processed.
// Per-item outcomes carry the real status, because a provider retrying an
// entire batch should not see a 4xx/5xx just because one event in it was
// malformed — see the note on decoding below for how one bad element is
// kept from poisoning the rest of the batch.
func (srv *Server) HandleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode into []json.RawMessage first, then unmarshal each element on
	// its own. This is what keeps one malformed element in a 200-event
	// batch from failing the other 199 — a whole-array []Event decode
	// would abort on the first bad element.
	var raw []json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, `{"error":"body must be a JSON array of events"}`, http.StatusBadRequest)
		return
	}

	results := make([]EventResult, 0, len(raw))
	accepted, rejected := 0, 0

	for _, item := range raw {
		var e Event
		if err := json.Unmarshal(item, &e); err != nil {
			results = append(results, EventResult{Outcome: OutcomeInvalid, Reason: "malformed JSON object"})
			rejected++
			continue
		}
		res := srv.Store.Ingest(e)
		if res.Outcome == OutcomeStored || res.Outcome == OutcomeDuplicate {
			accepted++
		} else {
			rejected++
		}
		results = append(results, res)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accepted": accepted,
		"rejected": rejected,
		"results":  results,
	})
}

// HandleStats implements GET /campaigns/{campaign_id}/stats.
func (srv *Server) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "campaigns" || parts[2] != "stats" {
		http.NotFound(w, r)
		return
	}
	campaignID := parts[1]
	stats := srv.Store.StatsFor(campaignID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Routes returns a ready-to-serve mux with all handlers registered.
func (srv *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", srv.HandleEvents)
	mux.HandleFunc("/campaigns/", srv.HandleStats)
	return mux
}
