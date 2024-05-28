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
		field.UUID("id", uuid.New()),
		field.Bytes("picture"),
		field.String("picture_mime"),
		field.Float("confidence"),
	}
}

// Edges of the Invader.
func (Invader) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", Event.Type).Ref("invaders").Required(),
	}
}
