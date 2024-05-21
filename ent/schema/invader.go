package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Invader holds the schema definition for the Invader entity.
type Invader struct {
	ent.Schema
}

// Fields of the Invader.
func (Invader) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Must(uuid.NewUUID())),
		field.Bytes("picture"),
		field.Float("confidence"),
	}
}

// Edges of the Invader.
func (Invader) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event_id", Event.Type).Ref("invaders").Required(),
	}
}