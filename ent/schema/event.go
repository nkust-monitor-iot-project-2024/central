package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Event holds the schema definition for the Event entity.
type Event struct {
	ent.Schema
}

// Fields of the Event.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("event_id", uuid.New()),
		field.Enum("type").Values("invaded", "movement"),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return nil
}
