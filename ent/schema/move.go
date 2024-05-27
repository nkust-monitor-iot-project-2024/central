package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Move holds the schema definition for the Move entity.
type Move struct {
	ent.Schema
}

// Fields of the Move.
func (Move) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()),
		field.Float("cycle"),
	}
}

// Edges of the Move.
func (Move) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", Event.Type).Ref("moves").Required(),
	}
}
