// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent/event"
	"github.com/nkust-monitor-iot-project-2024/central/ent/movement"
)

// MovementCreate is the builder for creating a Movement entity.
type MovementCreate struct {
	config
	mutation *MovementMutation
	hooks    []Hook
}

// SetPicture sets the "picture" field.
func (mc *MovementCreate) SetPicture(b []byte) *MovementCreate {
	mc.mutation.SetPicture(b)
	return mc
}

// SetPictureMime sets the "picture_mime" field.
func (mc *MovementCreate) SetPictureMime(s string) *MovementCreate {
	mc.mutation.SetPictureMime(s)
	return mc
}

// SetID sets the "id" field.
func (mc *MovementCreate) SetID(u uuid.UUID) *MovementCreate {
	mc.mutation.SetID(u)
	return mc
}

// AddEventIDs adds the "event" edge to the Event entity by IDs.
func (mc *MovementCreate) AddEventIDs(ids ...uuid.UUID) *MovementCreate {
	mc.mutation.AddEventIDs(ids...)
	return mc
}

// AddEvent adds the "event" edges to the Event entity.
func (mc *MovementCreate) AddEvent(e ...*Event) *MovementCreate {
	ids := make([]uuid.UUID, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return mc.AddEventIDs(ids...)
}

// Mutation returns the MovementMutation object of the builder.
func (mc *MovementCreate) Mutation() *MovementMutation {
	return mc.mutation
}

// Save creates the Movement in the database.
func (mc *MovementCreate) Save(ctx context.Context) (*Movement, error) {
	return withHooks(ctx, mc.sqlSave, mc.mutation, mc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (mc *MovementCreate) SaveX(ctx context.Context) *Movement {
	v, err := mc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (mc *MovementCreate) Exec(ctx context.Context) error {
	_, err := mc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mc *MovementCreate) ExecX(ctx context.Context) {
	if err := mc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (mc *MovementCreate) check() error {
	if _, ok := mc.mutation.Picture(); !ok {
		return &ValidationError{Name: "picture", err: errors.New(`ent: missing required field "Movement.picture"`)}
	}
	if _, ok := mc.mutation.PictureMime(); !ok {
		return &ValidationError{Name: "picture_mime", err: errors.New(`ent: missing required field "Movement.picture_mime"`)}
	}
	if len(mc.mutation.EventIDs()) == 0 {
		return &ValidationError{Name: "event", err: errors.New(`ent: missing required edge "Movement.event"`)}
	}
	return nil
}

func (mc *MovementCreate) sqlSave(ctx context.Context) (*Movement, error) {
	if err := mc.check(); err != nil {
		return nil, err
	}
	_node, _spec := mc.createSpec()
	if err := sqlgraph.CreateNode(ctx, mc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	mc.mutation.id = &_node.ID
	mc.mutation.done = true
	return _node, nil
}

func (mc *MovementCreate) createSpec() (*Movement, *sqlgraph.CreateSpec) {
	var (
		_node = &Movement{config: mc.config}
		_spec = sqlgraph.NewCreateSpec(movement.Table, sqlgraph.NewFieldSpec(movement.FieldID, field.TypeUUID))
	)
	if id, ok := mc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := mc.mutation.Picture(); ok {
		_spec.SetField(movement.FieldPicture, field.TypeBytes, value)
		_node.Picture = value
	}
	if value, ok := mc.mutation.PictureMime(); ok {
		_spec.SetField(movement.FieldPictureMime, field.TypeString, value)
		_node.PictureMime = value
	}
	if nodes := mc.mutation.EventIDs(); len(nodes) > 0 {
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
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// MovementCreateBulk is the builder for creating many Movement entities in bulk.
type MovementCreateBulk struct {
	config
	err      error
	builders []*MovementCreate
}

// Save creates the Movement entities in the database.
func (mcb *MovementCreateBulk) Save(ctx context.Context) ([]*Movement, error) {
	if mcb.err != nil {
		return nil, mcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(mcb.builders))
	nodes := make([]*Movement, len(mcb.builders))
	mutators := make([]Mutator, len(mcb.builders))
	for i := range mcb.builders {
		func(i int, root context.Context) {
			builder := mcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*MovementMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, mcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, mcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, mcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (mcb *MovementCreateBulk) SaveX(ctx context.Context) []*Movement {
	v, err := mcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (mcb *MovementCreateBulk) Exec(ctx context.Context) error {
	_, err := mcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mcb *MovementCreateBulk) ExecX(ctx context.Context) {
	if err := mcb.Exec(ctx); err != nil {
		panic(err)
	}
}
