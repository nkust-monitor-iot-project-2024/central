// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent/move"
)

// Move is the model entity for the Move schema.
type Move struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Cycle holds the value of the "cycle" field.
	Cycle float64 `json:"cycle,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the MoveQuery when eager-loading is set.
	Edges        MoveEdges `json:"edges"`
	selectValues sql.SelectValues
}

// MoveEdges holds the relations/edges for other nodes in the graph.
type MoveEdges struct {
	// Event holds the value of the event edge.
	Event []*Event `json:"event,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// EventOrErr returns the Event value or an error if the edge
// was not loaded in eager-loading.
func (e MoveEdges) EventOrErr() ([]*Event, error) {
	if e.loadedTypes[0] {
		return e.Event, nil
	}
	return nil, &NotLoadedError{edge: "event"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Move) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case move.FieldCycle:
			values[i] = new(sql.NullFloat64)
		case move.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Move fields.
func (m *Move) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case move.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				m.ID = *value
			}
		case move.FieldCycle:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fmt.Errorf("unexpected type %T for field cycle", values[i])
			} else if value.Valid {
				m.Cycle = value.Float64
			}
		default:
			m.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Move.
// This includes values selected through modifiers, order, etc.
func (m *Move) Value(name string) (ent.Value, error) {
	return m.selectValues.Get(name)
}

// QueryEvent queries the "event" edge of the Move entity.
func (m *Move) QueryEvent() *EventQuery {
	return NewMoveClient(m.config).QueryEvent(m)
}

// Update returns a builder for updating this Move.
// Note that you need to call Move.Unwrap() before calling this method if this Move
// was returned from a transaction, and the transaction was committed or rolled back.
func (m *Move) Update() *MoveUpdateOne {
	return NewMoveClient(m.config).UpdateOne(m)
}

// Unwrap unwraps the Move entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (m *Move) Unwrap() *Move {
	_tx, ok := m.config.driver.(*txDriver)
	if !ok {
		panic("ent: Move is not a transactional entity")
	}
	m.config.driver = _tx.drv
	return m
}

// String implements the fmt.Stringer.
func (m *Move) String() string {
	var builder strings.Builder
	builder.WriteString("Move(")
	builder.WriteString(fmt.Sprintf("id=%v, ", m.ID))
	builder.WriteString("cycle=")
	builder.WriteString(fmt.Sprintf("%v", m.Cycle))
	builder.WriteByte(')')
	return builder.String()
}

// Moves is a parsable slice of Move.
type Moves []*Move
