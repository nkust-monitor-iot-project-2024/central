// Code generated by ent, DO NOT EDIT.

package event

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldLTE(FieldID, id))
}

// DeviceID applies equality check predicate on the "device_id" field. It's identical to DeviceIDEQ.
func DeviceID(v string) predicate.Event {
	return predicate.Event(sql.FieldEQ(FieldDeviceID, v))
}

// ParentEventID applies equality check predicate on the "parent_event_id" field. It's identical to ParentEventIDEQ.
func ParentEventID(v uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldEQ(FieldParentEventID, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.Event {
	return predicate.Event(sql.FieldEQ(FieldCreatedAt, v))
}

// TypeEQ applies the EQ predicate on the "type" field.
func TypeEQ(v Type) predicate.Event {
	return predicate.Event(sql.FieldEQ(FieldType, v))
}

// TypeNEQ applies the NEQ predicate on the "type" field.
func TypeNEQ(v Type) predicate.Event {
	return predicate.Event(sql.FieldNEQ(FieldType, v))
}

// TypeIn applies the In predicate on the "type" field.
func TypeIn(vs ...Type) predicate.Event {
	return predicate.Event(sql.FieldIn(FieldType, vs...))
}

// TypeNotIn applies the NotIn predicate on the "type" field.
func TypeNotIn(vs ...Type) predicate.Event {
	return predicate.Event(sql.FieldNotIn(FieldType, vs...))
}

// DeviceIDEQ applies the EQ predicate on the "device_id" field.
func DeviceIDEQ(v string) predicate.Event {
	return predicate.Event(sql.FieldEQ(FieldDeviceID, v))
}

// DeviceIDNEQ applies the NEQ predicate on the "device_id" field.
func DeviceIDNEQ(v string) predicate.Event {
	return predicate.Event(sql.FieldNEQ(FieldDeviceID, v))
}

// DeviceIDIn applies the In predicate on the "device_id" field.
func DeviceIDIn(vs ...string) predicate.Event {
	return predicate.Event(sql.FieldIn(FieldDeviceID, vs...))
}

// DeviceIDNotIn applies the NotIn predicate on the "device_id" field.
func DeviceIDNotIn(vs ...string) predicate.Event {
	return predicate.Event(sql.FieldNotIn(FieldDeviceID, vs...))
}

// DeviceIDGT applies the GT predicate on the "device_id" field.
func DeviceIDGT(v string) predicate.Event {
	return predicate.Event(sql.FieldGT(FieldDeviceID, v))
}

// DeviceIDGTE applies the GTE predicate on the "device_id" field.
func DeviceIDGTE(v string) predicate.Event {
	return predicate.Event(sql.FieldGTE(FieldDeviceID, v))
}

// DeviceIDLT applies the LT predicate on the "device_id" field.
func DeviceIDLT(v string) predicate.Event {
	return predicate.Event(sql.FieldLT(FieldDeviceID, v))
}

// DeviceIDLTE applies the LTE predicate on the "device_id" field.
func DeviceIDLTE(v string) predicate.Event {
	return predicate.Event(sql.FieldLTE(FieldDeviceID, v))
}

// DeviceIDContains applies the Contains predicate on the "device_id" field.
func DeviceIDContains(v string) predicate.Event {
	return predicate.Event(sql.FieldContains(FieldDeviceID, v))
}

// DeviceIDHasPrefix applies the HasPrefix predicate on the "device_id" field.
func DeviceIDHasPrefix(v string) predicate.Event {
	return predicate.Event(sql.FieldHasPrefix(FieldDeviceID, v))
}

// DeviceIDHasSuffix applies the HasSuffix predicate on the "device_id" field.
func DeviceIDHasSuffix(v string) predicate.Event {
	return predicate.Event(sql.FieldHasSuffix(FieldDeviceID, v))
}

// DeviceIDEqualFold applies the EqualFold predicate on the "device_id" field.
func DeviceIDEqualFold(v string) predicate.Event {
	return predicate.Event(sql.FieldEqualFold(FieldDeviceID, v))
}

// DeviceIDContainsFold applies the ContainsFold predicate on the "device_id" field.
func DeviceIDContainsFold(v string) predicate.Event {
	return predicate.Event(sql.FieldContainsFold(FieldDeviceID, v))
}

// ParentEventIDEQ applies the EQ predicate on the "parent_event_id" field.
func ParentEventIDEQ(v uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldEQ(FieldParentEventID, v))
}

// ParentEventIDNEQ applies the NEQ predicate on the "parent_event_id" field.
func ParentEventIDNEQ(v uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldNEQ(FieldParentEventID, v))
}

// ParentEventIDIn applies the In predicate on the "parent_event_id" field.
func ParentEventIDIn(vs ...uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldIn(FieldParentEventID, vs...))
}

// ParentEventIDNotIn applies the NotIn predicate on the "parent_event_id" field.
func ParentEventIDNotIn(vs ...uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldNotIn(FieldParentEventID, vs...))
}

// ParentEventIDGT applies the GT predicate on the "parent_event_id" field.
func ParentEventIDGT(v uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldGT(FieldParentEventID, v))
}

// ParentEventIDGTE applies the GTE predicate on the "parent_event_id" field.
func ParentEventIDGTE(v uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldGTE(FieldParentEventID, v))
}

// ParentEventIDLT applies the LT predicate on the "parent_event_id" field.
func ParentEventIDLT(v uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldLT(FieldParentEventID, v))
}

// ParentEventIDLTE applies the LTE predicate on the "parent_event_id" field.
func ParentEventIDLTE(v uuid.UUID) predicate.Event {
	return predicate.Event(sql.FieldLTE(FieldParentEventID, v))
}

// ParentEventIDIsNil applies the IsNil predicate on the "parent_event_id" field.
func ParentEventIDIsNil() predicate.Event {
	return predicate.Event(sql.FieldIsNull(FieldParentEventID))
}

// ParentEventIDNotNil applies the NotNil predicate on the "parent_event_id" field.
func ParentEventIDNotNil() predicate.Event {
	return predicate.Event(sql.FieldNotNull(FieldParentEventID))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.Event {
	return predicate.Event(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.Event {
	return predicate.Event(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.Event {
	return predicate.Event(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.Event {
	return predicate.Event(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.Event {
	return predicate.Event(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.Event {
	return predicate.Event(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.Event {
	return predicate.Event(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.Event {
	return predicate.Event(sql.FieldLTE(FieldCreatedAt, v))
}

// HasInvaders applies the HasEdge predicate on the "invaders" edge.
func HasInvaders() predicate.Event {
	return predicate.Event(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, InvadersTable, InvadersPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasInvadersWith applies the HasEdge predicate on the "invaders" edge with a given conditions (other predicates).
func HasInvadersWith(preds ...predicate.Invader) predicate.Event {
	return predicate.Event(func(s *sql.Selector) {
		step := newInvadersStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasMovements applies the HasEdge predicate on the "movements" edge.
func HasMovements() predicate.Event {
	return predicate.Event(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, MovementsTable, MovementsPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasMovementsWith applies the HasEdge predicate on the "movements" edge with a given conditions (other predicates).
func HasMovementsWith(preds ...predicate.Movement) predicate.Event {
	return predicate.Event(func(s *sql.Selector) {
		step := newMovementsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasMoves applies the HasEdge predicate on the "moves" edge.
func HasMoves() predicate.Event {
	return predicate.Event(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, MovesTable, MovesPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasMovesWith applies the HasEdge predicate on the "moves" edge with a given conditions (other predicates).
func HasMovesWith(preds ...predicate.Move) predicate.Event {
	return predicate.Event(func(s *sql.Selector) {
		step := newMovesStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Event) predicate.Event {
	return predicate.Event(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Event) predicate.Event {
	return predicate.Event(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Event) predicate.Event {
	return predicate.Event(sql.NotPredicates(p))
}
