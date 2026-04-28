package domain

import "time"

type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeDenied  Outcome = "denied"
	OutcomeError   Outcome = "error"
)

type Event struct {
	ID           string            `json:"id"`
	OccurredAt   time.Time         `json:"occurred_at"`
	ActorType    string            `json:"actor_type"`
	ActorID      string            `json:"actor_id"`
	Action       string            `json:"action"`
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id,omitempty"`
	Outcome      Outcome           `json:"outcome"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}
