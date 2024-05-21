package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Movement holds the schema definition for the Movement entity.
type Movement struct {
	ent.Schema
}

// Fields of the Movement.
func (Movement) Fields() []ent.Field {
	return []ent.Field{
		field.Bytes("picture"),
	}
}

// Edges of the Movement.
func (Movement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("event_id", Event.Type),
	}
}
