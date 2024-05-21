package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Movement holds the schema definition for the Movement entity.
type Movement struct {
	ent.Schema
}

// Fields of the Movement.
func (Movement) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Must(uuid.NewUUID())),
		field.Bytes("picture"),
	}
}

// Edges of the Movement.
func (Movement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event_id", Event.Type).Ref("movements").Required(),
	}
}
