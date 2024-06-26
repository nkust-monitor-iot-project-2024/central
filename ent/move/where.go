// Code generated by ent, DO NOT EDIT.

package move

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.Move {
	return predicate.Move(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.Move {
	return predicate.Move(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.Move {
	return predicate.Move(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.Move {
	return predicate.Move(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.Move {
	return predicate.Move(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.Move {
	return predicate.Move(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.Move {
	return predicate.Move(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.Move {
	return predicate.Move(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.Move {
	return predicate.Move(sql.FieldLTE(FieldID, id))
}

// Cycle applies equality check predicate on the "cycle" field. It's identical to CycleEQ.
func Cycle(v float64) predicate.Move {
	return predicate.Move(sql.FieldEQ(FieldCycle, v))
}

// CycleEQ applies the EQ predicate on the "cycle" field.
func CycleEQ(v float64) predicate.Move {
	return predicate.Move(sql.FieldEQ(FieldCycle, v))
}

// CycleNEQ applies the NEQ predicate on the "cycle" field.
func CycleNEQ(v float64) predicate.Move {
	return predicate.Move(sql.FieldNEQ(FieldCycle, v))
}

// CycleIn applies the In predicate on the "cycle" field.
func CycleIn(vs ...float64) predicate.Move {
	return predicate.Move(sql.FieldIn(FieldCycle, vs...))
}

// CycleNotIn applies the NotIn predicate on the "cycle" field.
func CycleNotIn(vs ...float64) predicate.Move {
	return predicate.Move(sql.FieldNotIn(FieldCycle, vs...))
}

// CycleGT applies the GT predicate on the "cycle" field.
func CycleGT(v float64) predicate.Move {
	return predicate.Move(sql.FieldGT(FieldCycle, v))
}

// CycleGTE applies the GTE predicate on the "cycle" field.
func CycleGTE(v float64) predicate.Move {
	return predicate.Move(sql.FieldGTE(FieldCycle, v))
}

// CycleLT applies the LT predicate on the "cycle" field.
func CycleLT(v float64) predicate.Move {
	return predicate.Move(sql.FieldLT(FieldCycle, v))
}

// CycleLTE applies the LTE predicate on the "cycle" field.
func CycleLTE(v float64) predicate.Move {
	return predicate.Move(sql.FieldLTE(FieldCycle, v))
}

// HasEvent applies the HasEdge predicate on the "event" edge.
func HasEvent() predicate.Move {
	return predicate.Move(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, EventTable, EventPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEventWith applies the HasEdge predicate on the "event" edge with a given conditions (other predicates).
func HasEventWith(preds ...predicate.Event) predicate.Move {
	return predicate.Move(func(s *sql.Selector) {
		step := newEventStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Move) predicate.Move {
	return predicate.Move(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Move) predicate.Move {
	return predicate.Move(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Move) predicate.Move {
	return predicate.Move(sql.NotPredicates(p))
}
