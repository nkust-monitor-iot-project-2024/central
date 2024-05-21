// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// EventsColumns holds the columns for the "events" table.
	EventsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "type", Type: field.TypeEnum, Enums: []string{"invaded", "movement", "move"}},
		{Name: "created_at", Type: field.TypeTime},
	}
	// EventsTable holds the schema information for the "events" table.
	EventsTable = &schema.Table{
		Name:       "events",
		Columns:    EventsColumns,
		PrimaryKey: []*schema.Column{EventsColumns[0]},
	}
	// InvadersColumns holds the columns for the "invaders" table.
	InvadersColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "picture", Type: field.TypeBytes},
		{Name: "confidence", Type: field.TypeFloat64},
	}
	// InvadersTable holds the schema information for the "invaders" table.
	InvadersTable = &schema.Table{
		Name:       "invaders",
		Columns:    InvadersColumns,
		PrimaryKey: []*schema.Column{InvadersColumns[0]},
	}
	// MovesColumns holds the columns for the "moves" table.
	MovesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "cycle", Type: field.TypeFloat64},
	}
	// MovesTable holds the schema information for the "moves" table.
	MovesTable = &schema.Table{
		Name:       "moves",
		Columns:    MovesColumns,
		PrimaryKey: []*schema.Column{MovesColumns[0]},
	}
	// MovementsColumns holds the columns for the "movements" table.
	MovementsColumns = []*schema.Column{
		{Name: "id", Type: field.TypeUUID},
		{Name: "picture", Type: field.TypeBytes},
	}
	// MovementsTable holds the schema information for the "movements" table.
	MovementsTable = &schema.Table{
		Name:       "movements",
		Columns:    MovementsColumns,
		PrimaryKey: []*schema.Column{MovementsColumns[0]},
	}
	// EventInvadersColumns holds the columns for the "event_invaders" table.
	EventInvadersColumns = []*schema.Column{
		{Name: "event_id", Type: field.TypeUUID},
		{Name: "invader_id", Type: field.TypeUUID},
	}
	// EventInvadersTable holds the schema information for the "event_invaders" table.
	EventInvadersTable = &schema.Table{
		Name:       "event_invaders",
		Columns:    EventInvadersColumns,
		PrimaryKey: []*schema.Column{EventInvadersColumns[0], EventInvadersColumns[1]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "event_invaders_event_id",
				Columns:    []*schema.Column{EventInvadersColumns[0]},
				RefColumns: []*schema.Column{EventsColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "event_invaders_invader_id",
				Columns:    []*schema.Column{EventInvadersColumns[1]},
				RefColumns: []*schema.Column{InvadersColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// EventMovementsColumns holds the columns for the "event_movements" table.
	EventMovementsColumns = []*schema.Column{
		{Name: "event_id", Type: field.TypeUUID},
		{Name: "movement_id", Type: field.TypeUUID},
	}
	// EventMovementsTable holds the schema information for the "event_movements" table.
	EventMovementsTable = &schema.Table{
		Name:       "event_movements",
		Columns:    EventMovementsColumns,
		PrimaryKey: []*schema.Column{EventMovementsColumns[0], EventMovementsColumns[1]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "event_movements_event_id",
				Columns:    []*schema.Column{EventMovementsColumns[0]},
				RefColumns: []*schema.Column{EventsColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "event_movements_movement_id",
				Columns:    []*schema.Column{EventMovementsColumns[1]},
				RefColumns: []*schema.Column{MovementsColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// EventMovesColumns holds the columns for the "event_moves" table.
	EventMovesColumns = []*schema.Column{
		{Name: "event_id", Type: field.TypeUUID},
		{Name: "move_id", Type: field.TypeUUID},
	}
	// EventMovesTable holds the schema information for the "event_moves" table.
	EventMovesTable = &schema.Table{
		Name:       "event_moves",
		Columns:    EventMovesColumns,
		PrimaryKey: []*schema.Column{EventMovesColumns[0], EventMovesColumns[1]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "event_moves_event_id",
				Columns:    []*schema.Column{EventMovesColumns[0]},
				RefColumns: []*schema.Column{EventsColumns[0]},
				OnDelete:   schema.Cascade,
			},
			{
				Symbol:     "event_moves_move_id",
				Columns:    []*schema.Column{EventMovesColumns[1]},
				RefColumns: []*schema.Column{MovesColumns[0]},
				OnDelete:   schema.Cascade,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		EventsTable,
		InvadersTable,
		MovesTable,
		MovementsTable,
		EventInvadersTable,
		EventMovementsTable,
		EventMovesTable,
	}
)

func init() {
	EventInvadersTable.ForeignKeys[0].RefTable = EventsTable
	EventInvadersTable.ForeignKeys[1].RefTable = InvadersTable
	EventMovementsTable.ForeignKeys[0].RefTable = EventsTable
	EventMovementsTable.ForeignKeys[1].RefTable = MovementsTable
	EventMovesTable.ForeignKeys[0].RefTable = EventsTable
	EventMovesTable.ForeignKeys[1].RefTable = MovesTable
}
