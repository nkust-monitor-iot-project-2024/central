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
	"github.com/nkust-monitor-iot-project-2024/central/ent/movement"
	"github.com/nkust-monitor-iot-project-2024/central/ent/predicate"
)

// MovementUpdate is the builder for updating Movement entities.
type MovementUpdate struct {
	config
	hooks    []Hook
	mutation *MovementMutation
}

// Where appends a list predicates to the MovementUpdate builder.
func (mu *MovementUpdate) Where(ps ...predicate.Movement) *MovementUpdate {
	mu.mutation.Where(ps...)
	return mu
}

// SetPicture sets the "picture" field.
func (mu *MovementUpdate) SetPicture(b []byte) *MovementUpdate {
	mu.mutation.SetPicture(b)
	return mu
}

// SetPictureMime sets the "picture_mime" field.
func (mu *MovementUpdate) SetPictureMime(s string) *MovementUpdate {
	mu.mutation.SetPictureMime(s)
	return mu
}

// SetNillablePictureMime sets the "picture_mime" field if the given value is not nil.
func (mu *MovementUpdate) SetNillablePictureMime(s *string) *MovementUpdate {
	if s != nil {
		mu.SetPictureMime(*s)
	}
	return mu
}

// AddEventIDs adds the "event" edge to the Event entity by IDs.
func (mu *MovementUpdate) AddEventIDs(ids ...uuid.UUID) *MovementUpdate {
	mu.mutation.AddEventIDs(ids...)
	return mu
}

// AddEvent adds the "event" edges to the Event entity.
func (mu *MovementUpdate) AddEvent(e ...*Event) *MovementUpdate {
	ids := make([]uuid.UUID, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return mu.AddEventIDs(ids...)
}

// Mutation returns the MovementMutation object of the builder.
func (mu *MovementUpdate) Mutation() *MovementMutation {
	return mu.mutation
}

// ClearEvent clears all "event" edges to the Event entity.
func (mu *MovementUpdate) ClearEvent() *MovementUpdate {
	mu.mutation.ClearEvent()
	return mu
}

// RemoveEventIDs removes the "event" edge to Event entities by IDs.
func (mu *MovementUpdate) RemoveEventIDs(ids ...uuid.UUID) *MovementUpdate {
	mu.mutation.RemoveEventIDs(ids...)
	return mu
}

// RemoveEvent removes "event" edges to Event entities.
func (mu *MovementUpdate) RemoveEvent(e ...*Event) *MovementUpdate {
	ids := make([]uuid.UUID, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return mu.RemoveEventIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (mu *MovementUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, mu.sqlSave, mu.mutation, mu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (mu *MovementUpdate) SaveX(ctx context.Context) int {
	affected, err := mu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (mu *MovementUpdate) Exec(ctx context.Context) error {
	_, err := mu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mu *MovementUpdate) ExecX(ctx context.Context) {
	if err := mu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (mu *MovementUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(movement.Table, movement.Columns, sqlgraph.NewFieldSpec(movement.FieldID, field.TypeUUID))
	if ps := mu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := mu.mutation.Picture(); ok {
		_spec.SetField(movement.FieldPicture, field.TypeBytes, value)
	}
	if value, ok := mu.mutation.PictureMime(); ok {
		_spec.SetField(movement.FieldPictureMime, field.TypeString, value)
	}
	if mu.mutation.EventCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   movement.EventTable,
			Columns: movement.EventPrimaryKey,
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
			Table:   movement.EventTable,
			Columns: movement.EventPrimaryKey,
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
			Table:   movement.EventTable,
			Columns: movement.EventPrimaryKey,
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
			err = &NotFoundError{movement.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	mu.mutation.done = true
	return n, nil
}

// MovementUpdateOne is the builder for updating a single Movement entity.
type MovementUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *MovementMutation
}

// SetPicture sets the "picture" field.
func (muo *MovementUpdateOne) SetPicture(b []byte) *MovementUpdateOne {
	muo.mutation.SetPicture(b)
	return muo
}

// SetPictureMime sets the "picture_mime" field.
func (muo *MovementUpdateOne) SetPictureMime(s string) *MovementUpdateOne {
	muo.mutation.SetPictureMime(s)
	return muo
}

// SetNillablePictureMime sets the "picture_mime" field if the given value is not nil.
func (muo *MovementUpdateOne) SetNillablePictureMime(s *string) *MovementUpdateOne {
	if s != nil {
		muo.SetPictureMime(*s)
	}
	return muo
}

// AddEventIDs adds the "event" edge to the Event entity by IDs.
func (muo *MovementUpdateOne) AddEventIDs(ids ...uuid.UUID) *MovementUpdateOne {
	muo.mutation.AddEventIDs(ids...)
	return muo
}

// AddEvent adds the "event" edges to the Event entity.
func (muo *MovementUpdateOne) AddEvent(e ...*Event) *MovementUpdateOne {
	ids := make([]uuid.UUID, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return muo.AddEventIDs(ids...)
}

// Mutation returns the MovementMutation object of the builder.
func (muo *MovementUpdateOne) Mutation() *MovementMutation {
	return muo.mutation
}

// ClearEvent clears all "event" edges to the Event entity.
func (muo *MovementUpdateOne) ClearEvent() *MovementUpdateOne {
	muo.mutation.ClearEvent()
	return muo
}

// RemoveEventIDs removes the "event" edge to Event entities by IDs.
func (muo *MovementUpdateOne) RemoveEventIDs(ids ...uuid.UUID) *MovementUpdateOne {
	muo.mutation.RemoveEventIDs(ids...)
	return muo
}

// RemoveEvent removes "event" edges to Event entities.
func (muo *MovementUpdateOne) RemoveEvent(e ...*Event) *MovementUpdateOne {
	ids := make([]uuid.UUID, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return muo.RemoveEventIDs(ids...)
}

// Where appends a list predicates to the MovementUpdate builder.
func (muo *MovementUpdateOne) Where(ps ...predicate.Movement) *MovementUpdateOne {
	muo.mutation.Where(ps...)
	return muo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (muo *MovementUpdateOne) Select(field string, fields ...string) *MovementUpdateOne {
	muo.fields = append([]string{field}, fields...)
	return muo
}

// Save executes the query and returns the updated Movement entity.
func (muo *MovementUpdateOne) Save(ctx context.Context) (*Movement, error) {
	return withHooks(ctx, muo.sqlSave, muo.mutation, muo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (muo *MovementUpdateOne) SaveX(ctx context.Context) *Movement {
	node, err := muo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (muo *MovementUpdateOne) Exec(ctx context.Context) error {
	_, err := muo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (muo *MovementUpdateOne) ExecX(ctx context.Context) {
	if err := muo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (muo *MovementUpdateOne) sqlSave(ctx context.Context) (_node *Movement, err error) {
	_spec := sqlgraph.NewUpdateSpec(movement.Table, movement.Columns, sqlgraph.NewFieldSpec(movement.FieldID, field.TypeUUID))
	id, ok := muo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Movement.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := muo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, movement.FieldID)
		for _, f := range fields {
			if !movement.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != movement.FieldID {
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
	if value, ok := muo.mutation.Picture(); ok {
		_spec.SetField(movement.FieldPicture, field.TypeBytes, value)
	}
	if value, ok := muo.mutation.PictureMime(); ok {
		_spec.SetField(movement.FieldPictureMime, field.TypeString, value)
	}
	if muo.mutation.EventCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   movement.EventTable,
			Columns: movement.EventPrimaryKey,
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
			Table:   movement.EventTable,
			Columns: movement.EventPrimaryKey,
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
			Table:   movement.EventTable,
			Columns: movement.EventPrimaryKey,
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
	_node = &Movement{config: muo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, muo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{movement.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	muo.mutation.done = true
	return _node, nil
}
