// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent/event"
	"github.com/nkust-monitor-iot-project-2024/central/ent/invader"
	"github.com/nkust-monitor-iot-project-2024/central/ent/movement"
	"github.com/nkust-monitor-iot-project-2024/central/ent/predicate"
)

// EventUpdate is the builder for updating Event entities.
type EventUpdate struct {
	config
	hooks    []Hook
	mutation *EventMutation
}

// Where appends a list predicates to the EventUpdate builder.
func (eu *EventUpdate) Where(ps ...predicate.Event) *EventUpdate {
	eu.mutation.Where(ps...)
	return eu
}

// SetType sets the "type" field.
func (eu *EventUpdate) SetType(e event.Type) *EventUpdate {
	eu.mutation.SetType(e)
	return eu
}

// SetNillableType sets the "type" field if the given value is not nil.
func (eu *EventUpdate) SetNillableType(e *event.Type) *EventUpdate {
	if e != nil {
		eu.SetType(*e)
	}
	return eu
}

// SetCreatedAt sets the "created_at" field.
func (eu *EventUpdate) SetCreatedAt(t time.Time) *EventUpdate {
	eu.mutation.SetCreatedAt(t)
	return eu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (eu *EventUpdate) SetNillableCreatedAt(t *time.Time) *EventUpdate {
	if t != nil {
		eu.SetCreatedAt(*t)
	}
	return eu
}

// AddInvaderIDs adds the "invaders" edge to the Invader entity by IDs.
func (eu *EventUpdate) AddInvaderIDs(ids ...uuid.UUID) *EventUpdate {
	eu.mutation.AddInvaderIDs(ids...)
	return eu
}

// AddInvaders adds the "invaders" edges to the Invader entity.
func (eu *EventUpdate) AddInvaders(i ...*Invader) *EventUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return eu.AddInvaderIDs(ids...)
}

// AddMovementIDs adds the "movements" edge to the Movement entity by IDs.
func (eu *EventUpdate) AddMovementIDs(ids ...uuid.UUID) *EventUpdate {
	eu.mutation.AddMovementIDs(ids...)
	return eu
}

// AddMovements adds the "movements" edges to the Movement entity.
func (eu *EventUpdate) AddMovements(m ...*Movement) *EventUpdate {
	ids := make([]uuid.UUID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return eu.AddMovementIDs(ids...)
}

// Mutation returns the EventMutation object of the builder.
func (eu *EventUpdate) Mutation() *EventMutation {
	return eu.mutation
}

// ClearInvaders clears all "invaders" edges to the Invader entity.
func (eu *EventUpdate) ClearInvaders() *EventUpdate {
	eu.mutation.ClearInvaders()
	return eu
}

// RemoveInvaderIDs removes the "invaders" edge to Invader entities by IDs.
func (eu *EventUpdate) RemoveInvaderIDs(ids ...uuid.UUID) *EventUpdate {
	eu.mutation.RemoveInvaderIDs(ids...)
	return eu
}

// RemoveInvaders removes "invaders" edges to Invader entities.
func (eu *EventUpdate) RemoveInvaders(i ...*Invader) *EventUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return eu.RemoveInvaderIDs(ids...)
}

// ClearMovements clears all "movements" edges to the Movement entity.
func (eu *EventUpdate) ClearMovements() *EventUpdate {
	eu.mutation.ClearMovements()
	return eu
}

// RemoveMovementIDs removes the "movements" edge to Movement entities by IDs.
func (eu *EventUpdate) RemoveMovementIDs(ids ...uuid.UUID) *EventUpdate {
	eu.mutation.RemoveMovementIDs(ids...)
	return eu
}

// RemoveMovements removes "movements" edges to Movement entities.
func (eu *EventUpdate) RemoveMovements(m ...*Movement) *EventUpdate {
	ids := make([]uuid.UUID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return eu.RemoveMovementIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (eu *EventUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, eu.sqlSave, eu.mutation, eu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (eu *EventUpdate) SaveX(ctx context.Context) int {
	affected, err := eu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (eu *EventUpdate) Exec(ctx context.Context) error {
	_, err := eu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (eu *EventUpdate) ExecX(ctx context.Context) {
	if err := eu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (eu *EventUpdate) check() error {
	if v, ok := eu.mutation.GetType(); ok {
		if err := event.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "Event.type": %w`, err)}
		}
	}
	return nil
}

func (eu *EventUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := eu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(event.Table, event.Columns, sqlgraph.NewFieldSpec(event.FieldID, field.TypeUUID))
	if ps := eu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := eu.mutation.GetType(); ok {
		_spec.SetField(event.FieldType, field.TypeEnum, value)
	}
	if value, ok := eu.mutation.CreatedAt(); ok {
		_spec.SetField(event.FieldCreatedAt, field.TypeTime, value)
	}
	if eu.mutation.InvadersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.InvadersTable,
			Columns: event.InvadersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(invader.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.RemovedInvadersIDs(); len(nodes) > 0 && !eu.mutation.InvadersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.InvadersTable,
			Columns: event.InvadersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(invader.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.InvadersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.InvadersTable,
			Columns: event.InvadersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(invader.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if eu.mutation.MovementsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.MovementsTable,
			Columns: event.MovementsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(movement.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.RemovedMovementsIDs(); len(nodes) > 0 && !eu.mutation.MovementsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.MovementsTable,
			Columns: event.MovementsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(movement.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.MovementsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.MovementsTable,
			Columns: event.MovementsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(movement.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, eu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{event.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	eu.mutation.done = true
	return n, nil
}

// EventUpdateOne is the builder for updating a single Event entity.
type EventUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *EventMutation
}

// SetType sets the "type" field.
func (euo *EventUpdateOne) SetType(e event.Type) *EventUpdateOne {
	euo.mutation.SetType(e)
	return euo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (euo *EventUpdateOne) SetNillableType(e *event.Type) *EventUpdateOne {
	if e != nil {
		euo.SetType(*e)
	}
	return euo
}

// SetCreatedAt sets the "created_at" field.
func (euo *EventUpdateOne) SetCreatedAt(t time.Time) *EventUpdateOne {
	euo.mutation.SetCreatedAt(t)
	return euo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (euo *EventUpdateOne) SetNillableCreatedAt(t *time.Time) *EventUpdateOne {
	if t != nil {
		euo.SetCreatedAt(*t)
	}
	return euo
}

// AddInvaderIDs adds the "invaders" edge to the Invader entity by IDs.
func (euo *EventUpdateOne) AddInvaderIDs(ids ...uuid.UUID) *EventUpdateOne {
	euo.mutation.AddInvaderIDs(ids...)
	return euo
}

// AddInvaders adds the "invaders" edges to the Invader entity.
func (euo *EventUpdateOne) AddInvaders(i ...*Invader) *EventUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return euo.AddInvaderIDs(ids...)
}

// AddMovementIDs adds the "movements" edge to the Movement entity by IDs.
func (euo *EventUpdateOne) AddMovementIDs(ids ...uuid.UUID) *EventUpdateOne {
	euo.mutation.AddMovementIDs(ids...)
	return euo
}

// AddMovements adds the "movements" edges to the Movement entity.
func (euo *EventUpdateOne) AddMovements(m ...*Movement) *EventUpdateOne {
	ids := make([]uuid.UUID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return euo.AddMovementIDs(ids...)
}

// Mutation returns the EventMutation object of the builder.
func (euo *EventUpdateOne) Mutation() *EventMutation {
	return euo.mutation
}

// ClearInvaders clears all "invaders" edges to the Invader entity.
func (euo *EventUpdateOne) ClearInvaders() *EventUpdateOne {
	euo.mutation.ClearInvaders()
	return euo
}

// RemoveInvaderIDs removes the "invaders" edge to Invader entities by IDs.
func (euo *EventUpdateOne) RemoveInvaderIDs(ids ...uuid.UUID) *EventUpdateOne {
	euo.mutation.RemoveInvaderIDs(ids...)
	return euo
}

// RemoveInvaders removes "invaders" edges to Invader entities.
func (euo *EventUpdateOne) RemoveInvaders(i ...*Invader) *EventUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return euo.RemoveInvaderIDs(ids...)
}

// ClearMovements clears all "movements" edges to the Movement entity.
func (euo *EventUpdateOne) ClearMovements() *EventUpdateOne {
	euo.mutation.ClearMovements()
	return euo
}

// RemoveMovementIDs removes the "movements" edge to Movement entities by IDs.
func (euo *EventUpdateOne) RemoveMovementIDs(ids ...uuid.UUID) *EventUpdateOne {
	euo.mutation.RemoveMovementIDs(ids...)
	return euo
}

// RemoveMovements removes "movements" edges to Movement entities.
func (euo *EventUpdateOne) RemoveMovements(m ...*Movement) *EventUpdateOne {
	ids := make([]uuid.UUID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return euo.RemoveMovementIDs(ids...)
}

// Where appends a list predicates to the EventUpdate builder.
func (euo *EventUpdateOne) Where(ps ...predicate.Event) *EventUpdateOne {
	euo.mutation.Where(ps...)
	return euo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (euo *EventUpdateOne) Select(field string, fields ...string) *EventUpdateOne {
	euo.fields = append([]string{field}, fields...)
	return euo
}

// Save executes the query and returns the updated Event entity.
func (euo *EventUpdateOne) Save(ctx context.Context) (*Event, error) {
	return withHooks(ctx, euo.sqlSave, euo.mutation, euo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (euo *EventUpdateOne) SaveX(ctx context.Context) *Event {
	node, err := euo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (euo *EventUpdateOne) Exec(ctx context.Context) error {
	_, err := euo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (euo *EventUpdateOne) ExecX(ctx context.Context) {
	if err := euo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (euo *EventUpdateOne) check() error {
	if v, ok := euo.mutation.GetType(); ok {
		if err := event.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "Event.type": %w`, err)}
		}
	}
	return nil
}

func (euo *EventUpdateOne) sqlSave(ctx context.Context) (_node *Event, err error) {
	if err := euo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(event.Table, event.Columns, sqlgraph.NewFieldSpec(event.FieldID, field.TypeUUID))
	id, ok := euo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Event.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := euo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, event.FieldID)
		for _, f := range fields {
			if !event.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != event.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := euo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := euo.mutation.GetType(); ok {
		_spec.SetField(event.FieldType, field.TypeEnum, value)
	}
	if value, ok := euo.mutation.CreatedAt(); ok {
		_spec.SetField(event.FieldCreatedAt, field.TypeTime, value)
	}
	if euo.mutation.InvadersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.InvadersTable,
			Columns: event.InvadersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(invader.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.RemovedInvadersIDs(); len(nodes) > 0 && !euo.mutation.InvadersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.InvadersTable,
			Columns: event.InvadersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(invader.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.InvadersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.InvadersTable,
			Columns: event.InvadersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(invader.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if euo.mutation.MovementsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.MovementsTable,
			Columns: event.MovementsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(movement.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.RemovedMovementsIDs(); len(nodes) > 0 && !euo.mutation.MovementsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.MovementsTable,
			Columns: event.MovementsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(movement.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.MovementsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   event.MovementsTable,
			Columns: event.MovementsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(movement.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Event{config: euo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, euo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{event.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	euo.mutation.done = true
	return _node, nil
}