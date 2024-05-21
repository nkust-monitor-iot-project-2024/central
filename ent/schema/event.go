package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
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
		field.UUID("id", uuid.Must(uuid.NewV7())),
		field.Enum("type").Values("invaded", "movement", "move"),
		field.String("device_id"),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("invaders", Invader.Type),
		edge.To("movements", Movement.Type),
		edge.To("moves", Move.Type),
	}
}
