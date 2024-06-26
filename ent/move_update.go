// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent/event"
	"github.com/nkust-monitor-iot-project-2024/central/ent/move"
	"github.com/nkust-monitor-iot-project-2024/central/ent/predicate"
)

// MoveUpdate is the builder for updating Move entities.
type MoveUpdate struct {
	config
	hooks    []Hook
	mutation *MoveMutation
}

// Where appends a list predicates to the MoveUpdate builder.
func (mu *MoveUpdate) Where(ps ...predicate.Move) *MoveUpdate {
	mu.mutation.Where(ps...)
	return mu
}

// SetCycle sets the "cycle" field.
func (mu *MoveUpdate) SetCycle(f float64) *MoveUpdate {
	mu.mutation.ResetCycle()
	mu.mutation.SetCycle(f)
	return mu
}

// SetNillableCycle sets the "cycle" field if the given value is not nil.
func (mu *MoveUpdate) SetNillableCycle(f *float64) *MoveUpdate {
	if f != nil {
		mu.SetCycle(*f)
	}
	return mu
}

// AddCycle adds f to the "cycle" field.
func (mu *MoveUpdate) AddCycle(f float64) *MoveUpdate {
	mu.mutation.AddCycle(f)
	return mu
}

// AddEventIDs adds the "event" edge to the Event entity by IDs.
func (mu *MoveUpdate) AddEventIDs(ids ...uuid.UUID) *MoveUpdate {
	mu.mutation.AddEventIDs(ids...)
	return mu
}

// AddEvent adds the "event" edges to the Event entity.
func (mu *MoveUpdate) AddEvent(e ...*Event) *MoveUpdate {
	ids := make([]uuid.UUID, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return mu.AddEventIDs(ids...)
}

// Mutation returns the MoveMutation object of the builder.
func (mu *MoveUpdate) Mutation() *MoveMutation {
	return mu.mutation
}

// ClearEvent clears all "event" edges to the Event entity.
func (mu *MoveUpdate) ClearEvent() *MoveUpdate {
	mu.mutation.ClearEvent()
	return mu
}

// RemoveEventIDs removes the "event" edge to Event entities by IDs.
func (mu *MoveUpdate) RemoveEventIDs(ids ...uuid.UUID) *MoveUpdate {
	mu.mutation.RemoveEventIDs(ids...)
	return mu
}

// RemoveEvent removes "event" edges to Event entities.
func (mu *MoveUpdate) RemoveEvent(e ...*Event) *MoveUpdate {
	ids := make([]uuid.UUID, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return mu.RemoveEventIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (mu *MoveUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, mu.sqlSave, mu.mutation, mu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (mu *MoveUpdate) SaveX(ctx context.Context) int {
	affected, err := mu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (mu *MoveUpdate) Exec(ctx context.Context) error {
	_, err := mu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mu *MoveUpdate) ExecX(ctx context.Context) {
	if err := mu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (mu *MoveUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(move.Table, move.Columns, sqlgraph.NewFieldSpec(move.FieldID, field.TypeUUID))
	if ps := mu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := mu.mutation.Cycle(); ok {
		_spec.SetField(move.FieldCycle, field.TypeFloat64, value)
	}
	if value, ok := mu.mutation.AddedCycle(); ok {
		_spec.AddField(move.FieldCycle, field.TypeFloat64, value)
	}
	if mu.mutation.EventCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   move.EventTable,
			Columns: move.EventPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(event.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := mu.mutation.RemovedEventIDs(); len(nodes) > 0 && !mu.mutation.EventCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   move.EventTable,
			Columns: move.EventPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(event.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := mu.mutation.EventIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   move.EventTable,
			Columns: move.EventPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(event.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, mu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{move.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	mu.mutation.done = true
	return n, nil
}

// MoveUpdateOne is the builder for updating a single Move entity.
type MoveUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *MoveMutation
}

// SetCycle sets the "cycle" field.
func (muo *MoveUpdateOne) SetCycle(f float64) *MoveUpdateOne {
	muo.mutation.ResetCycle()
	muo.mutation.SetCycle(f)
	return muo
}

// SetNillableCycle sets the "cycle" field if the given value is not nil.
func (muo *MoveUpdateOne) SetNillableCycle(f *float64) *MoveUpdateOne {
	if f != nil {
		muo.SetCycle(*f)
	}
	return muo
}

// AddCycle adds f to the "cycle" field.
func (muo *MoveUpdateOne) AddCycle(f float64) *MoveUpdateOne {
	muo.mutation.AddCycle(f)
	return muo
}

// AddEventIDs adds the "event" edge to the Event entity by IDs.
func (muo *MoveUpdateOne) AddEventIDs(ids ...uuid.UUID) *MoveUpdateOne {
	muo.mutation.AddEventIDs(ids...)
	return muo
}

// AddEvent adds the "event" edges to the Event entity.
func (muo *MoveUpdateOne) AddEvent(e ...*Event) *MoveUpdateOne {
	ids := make([]uuid.UUID, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return muo.AddEventIDs(ids...)
}

// Mutation returns the MoveMutation object of the builder.
func (muo *MoveUpdateOne) Mutation() *MoveMutation {
	return muo.mutation
}

// ClearEvent clears all "event" edges to the Event entity.
func (muo *MoveUpdateOne) ClearEvent() *MoveUpdateOne {
	muo.mutation.ClearEvent()
	return muo
}

// RemoveEventIDs removes the "event" edge to Event entities by IDs.
func (muo *MoveUpdateOne) RemoveEventIDs(ids ...uuid.UUID) *MoveUpdateOne {
	muo.mutation.RemoveEventIDs(ids...)
	return muo
}

// RemoveEvent removes "event" edges to Event entities.
func (muo *MoveUpdateOne) RemoveEvent(e ...*Event) *MoveUpdateOne {
	ids := make([]uuid.UUID, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return muo.RemoveEventIDs(ids...)
}

// Where appends a list predicates to the MoveUpdate builder.
func (muo *MoveUpdateOne) Where(ps ...predicate.Move) *MoveUpdateOne {
	muo.mutation.Where(ps...)
	return muo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (muo *MoveUpdateOne) Select(field string, fields ...string) *MoveUpdateOne {
	muo.fields = append([]string{field}, fields...)
	return muo
}

// Save executes the query and returns the updated Move entity.
func (muo *MoveUpdateOne) Save(ctx context.Context) (*Move, error) {
	return withHooks(ctx, muo.sqlSave, muo.mutation, muo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (muo *MoveUpdateOne) SaveX(ctx context.Context) *Move {
	node, err := muo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (muo *MoveUpdateOne) Exec(ctx context.Context) error {
	_, err := muo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (muo *MoveUpdateOne) ExecX(ctx context.Context) {
	if err := muo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (muo *MoveUpdateOne) sqlSave(ctx context.Context) (_node *Move, err error) {
	_spec := sqlgraph.NewUpdateSpec(move.Table, move.Columns, sqlgraph.NewFieldSpec(move.FieldID, field.TypeUUID))
	id, ok := muo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Move.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := muo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, move.FieldID)
		for _, f := range fields {
			if !move.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != move.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := muo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := muo.mutation.Cycle(); ok {
		_spec.SetField(move.FieldCycle, field.TypeFloat64, value)
	}
	if value, ok := muo.mutation.AddedCycle(); ok {
		_spec.AddField(move.FieldCycle, field.TypeFloat64, value)
	}
	if muo.mutation.EventCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   move.EventTable,
			Columns: move.EventPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(event.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := muo.mutation.RemovedEventIDs(); len(nodes) > 0 && !muo.mutation.EventCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   move.EventTable,
			Columns: move.EventPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(event.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := muo.mutation.EventIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   move.EventTable,
			Columns: move.EventPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(event.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Move{config: muo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, muo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{move.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	muo.mutation.done = true
	return _node, nil
}
