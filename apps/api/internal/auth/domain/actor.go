package domain

type ActorType string

const (
	ActorUser      ActorType = "user"
	ActorService   ActorType = "service"
	ActorAnonymous ActorType = "anonymous"
)

type Actor struct {
	ID    string    `json:"id"`
	Type  ActorType `json:"type"`
	Roles []string  `json:"roles"`
}

func Anonymous() Actor {
	return Actor{ID: "anonymous", Type: ActorAnonymous}
}
