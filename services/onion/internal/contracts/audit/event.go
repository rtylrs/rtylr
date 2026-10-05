package audit

const MutationRecorded = "entity_mutated"

type Event struct {
	BusinessID  int64
	ActorID     int64
	EntityType  string
	EntityID    int64
	Action      string
	Description string
	OldValues   any
	NewValues   any
}
