package events

import (
	"fmt"
	"strings"
	"time"
)

// Event is the fixed provider webhook payload.
type Event struct {
	EventID    string                 `json:"event_id"`
	CampaignID string                 `json:"campaign_id"`
	ContactID  string                 `json:"contact_id"`
	Type       string                 `json:"type"`
	Timestamp  string                 `json:"timestamp"` // RFC3339
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

var validTypes = map[string]bool{
	"sent":      true,
	"delivered": true,
	"opened":    true,
	"clicked":   true,
}

// Validate checks the required fields of an event. It returns a
// human-readable reason and false if the event should be rejected.
func Validate(e Event) (reason string, ok bool) {
	if strings.TrimSpace(e.EventID) == "" {
		return "event_id is required", false
	}
	if strings.TrimSpace(e.CampaignID) == "" {
		return "campaign_id is required", false
	}
	if !validTypes[e.Type] {
		return fmt.Sprintf("type %q is not one of sent|delivered|opened|clicked", e.Type), false
	}
	if _, err := time.Parse(time.RFC3339, e.Timestamp); err != nil {
		return "timestamp is not valid RFC3339", false
	}
	return "", true
}

// SameEvent reports whether two events with the same event_id represent an
// identical retry (true) or a suspicious payload change on a reused
// event_id (false).
func SameEvent(a, b Event) bool {
	return a.CampaignID == b.CampaignID &&
		a.ContactID == b.ContactID &&
		a.Type == b.Type &&
		a.Timestamp == b.Timestamp
}

// Stats holds per-campaign counts. Counts are of distinct event_ids, so
// provider retries never inflate the numbers.
type Stats struct {
	CampaignID     string `json:"campaign_id"`
	Sent           int    `json:"sent"`
	Delivered      int    `json:"delivered"`
	Opened         int    `json:"opened"`
	Clicked        int    `json:"clicked"`
	UniqueContacts int    `json:"unique_contacts"`
	LastEventAt    string `json:"last_event_at,omitempty"`
}

// Outcome describes what happened to a single event on ingest.
type Outcome string

const (
	OutcomeStored    Outcome = "stored"
	OutcomeDuplicate Outcome = "duplicate"
	OutcomeConflict  Outcome = "conflict" // same event_id, different payload
	OutcomeInvalid   Outcome = "invalid"
)

// EventResult is the per-event outcome returned from a batch ingest.
type EventResult struct {
	EventID string  `json:"event_id,omitempty"`
	Outcome Outcome `json:"outcome"`
	Reason  string  `json:"reason,omitempty"`
}
