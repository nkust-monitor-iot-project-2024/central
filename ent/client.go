// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent/migrate"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/nkust-monitor-iot-project-2024/central/ent/event"
	"github.com/nkust-monitor-iot-project-2024/central/ent/invader"
	"github.com/nkust-monitor-iot-project-2024/central/ent/movement"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Event is the client for interacting with the Event builders.
	Event *EventClient
	// Invader is the client for interacting with the Invader builders.
	Invader *InvaderClient
	// Movement is the client for interacting with the Movement builders.
	Movement *MovementClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	client := &Client{config: newConfig(opts...)}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.Event = NewEventClient(c.config)
	c.Invader = NewInvaderClient(c.config)
	c.Movement = NewMovementClient(c.config)
}

type (
	// config is the configuration for the client and its builder.
	config struct {
		// driver used for executing database requests.
		driver dialect.Driver
		// debug enable a debug logging.
		debug bool
		// log used for logging on debug mode.
		log func(...any)
		// hooks to execute on mutations.
		hooks *hooks
		// interceptors to execute on queries.
		inters *inters
	}
	// Option function to configure the client.
	Option func(*config)
)

// newConfig creates a new config for the client.
func newConfig(opts ...Option) config {
	cfg := config{log: log.Println, hooks: &hooks{}, inters: &inters{}}
	cfg.options(opts...)
	return cfg
}

// options applies the options on the config object.
func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	if c.debug {
		c.driver = dialect.Debug(c.driver, c.log)
	}
}

// Debug enables debug logging on the ent.Driver.
func Debug() Option {
	return func(c *config) {
		c.debug = true
	}
}

// Log sets the logging function for debug mode.
func Log(fn func(...any)) Option {
	return func(c *config) {
		c.log = fn
	}
}

// Driver configures the client driver.
func Driver(driver dialect.Driver) Option {
	return func(c *config) {
		c.driver = driver
	}
}

// Open opens a database/sql.DB specified by the driver name and
// the data source name, and returns a new client attached to it.
// Optional parameters can be added for configuring the client.
func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return NewClient(append(options, Driver(drv))...), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}

// ErrTxStarted is returned when trying to start a new transaction from a transactional client.
var ErrTxStarted = errors.New("ent: cannot start a transaction within a transaction")

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, ErrTxStarted
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:      ctx,
		config:   cfg,
		Event:    NewEventClient(cfg),
		Invader:  NewInvaderClient(cfg),
		Movement: NewMovementClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := c.driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = &txDriver{tx: tx, drv: c.driver}
	return &Tx{
		ctx:      ctx,
		config:   cfg,
		Event:    NewEventClient(cfg),
		Invader:  NewInvaderClient(cfg),
		Movement: NewMovementClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Event.
//		Query().
//		Count(ctx)
func (c *Client) Debug() *Client {
	if c.debug {
		return c
	}
	cfg := c.config
	cfg.driver = dialect.Debug(c.driver, c.log)
	client := &Client{config: cfg}
	client.init()
	return client
}

// Close closes the database connection and prevents new queries from starting.
func (c *Client) Close() error {
	return c.driver.Close()
}

// Use adds the mutation hooks to all the entity clients.
// In order to add hooks to a specific client, call: `client.Node.Use(...)`.
func (c *Client) Use(hooks ...Hook) {
	c.Event.Use(hooks...)
	c.Invader.Use(hooks...)
	c.Movement.Use(hooks...)
}

// Intercept adds the query interceptors to all the entity clients.
// In order to add interceptors to a specific client, call: `client.Node.Intercept(...)`.
func (c *Client) Intercept(interceptors ...Interceptor) {
	c.Event.Intercept(interceptors...)
	c.Invader.Intercept(interceptors...)
	c.Movement.Intercept(interceptors...)
}

// Mutate implements the ent.Mutator interface.
func (c *Client) Mutate(ctx context.Context, m Mutation) (Value, error) {
	switch m := m.(type) {
	case *EventMutation:
		return c.Event.mutate(ctx, m)
	case *InvaderMutation:
		return c.Invader.mutate(ctx, m)
	case *MovementMutation:
		return c.Movement.mutate(ctx, m)
	default:
		return nil, fmt.Errorf("ent: unknown mutation type %T", m)
	}
}

// EventClient is a client for the Event schema.
type EventClient struct {
	config
}

// NewEventClient returns a client for the Event from the given config.
func NewEventClient(c config) *EventClient {
	return &EventClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `event.Hooks(f(g(h())))`.
func (c *EventClient) Use(hooks ...Hook) {
	c.hooks.Event = append(c.hooks.Event, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `event.Intercept(f(g(h())))`.
func (c *EventClient) Intercept(interceptors ...Interceptor) {
	c.inters.Event = append(c.inters.Event, interceptors...)
}

// Create returns a builder for creating a Event entity.
func (c *EventClient) Create() *EventCreate {
	mutation := newEventMutation(c.config, OpCreate)
	return &EventCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Event entities.
func (c *EventClient) CreateBulk(builders ...*EventCreate) *EventCreateBulk {
	return &EventCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *EventClient) MapCreateBulk(slice any, setFunc func(*EventCreate, int)) *EventCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &EventCreateBulk{err: fmt.Errorf("calling to EventClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*EventCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &EventCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Event.
func (c *EventClient) Update() *EventUpdate {
	mutation := newEventMutation(c.config, OpUpdate)
	return &EventUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *EventClient) UpdateOne(e *Event) *EventUpdateOne {
	mutation := newEventMutation(c.config, OpUpdateOne, withEvent(e))
	return &EventUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *EventClient) UpdateOneID(id uuid.UUID) *EventUpdateOne {
	mutation := newEventMutation(c.config, OpUpdateOne, withEventID(id))
	return &EventUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Event.
func (c *EventClient) Delete() *EventDelete {
	mutation := newEventMutation(c.config, OpDelete)
	return &EventDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *EventClient) DeleteOne(e *Event) *EventDeleteOne {
	return c.DeleteOneID(e.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *EventClient) DeleteOneID(id uuid.UUID) *EventDeleteOne {
	builder := c.Delete().Where(event.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &EventDeleteOne{builder}
}

// Query returns a query builder for Event.
func (c *EventClient) Query() *EventQuery {
	return &EventQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeEvent},
		inters: c.Interceptors(),
	}
}

// Get returns a Event entity by its id.
func (c *EventClient) Get(ctx context.Context, id uuid.UUID) (*Event, error) {
	return c.Query().Where(event.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *EventClient) GetX(ctx context.Context, id uuid.UUID) *Event {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryInvaders queries the invaders edge of a Event.
func (c *EventClient) QueryInvaders(e *Event) *InvaderQuery {
	query := (&InvaderClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := e.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(event.Table, event.FieldID, id),
			sqlgraph.To(invader.Table, invader.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, event.InvadersTable, event.InvadersPrimaryKey...),
		)
		fromV = sqlgraph.Neighbors(e.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// QueryMovements queries the movements edge of a Event.
func (c *EventClient) QueryMovements(e *Event) *MovementQuery {
	query := (&MovementClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := e.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(event.Table, event.FieldID, id),
			sqlgraph.To(movement.Table, movement.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, event.MovementsTable, event.MovementsPrimaryKey...),
		)
		fromV = sqlgraph.Neighbors(e.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *EventClient) Hooks() []Hook {
	return c.hooks.Event
}

// Interceptors returns the client interceptors.
func (c *EventClient) Interceptors() []Interceptor {
	return c.inters.Event
}

func (c *EventClient) mutate(ctx context.Context, m *EventMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&EventCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&EventUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&EventUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&EventDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Event mutation op: %q", m.Op())
	}
}

// InvaderClient is a client for the Invader schema.
type InvaderClient struct {
	config
}

// NewInvaderClient returns a client for the Invader from the given config.
func NewInvaderClient(c config) *InvaderClient {
	return &InvaderClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `invader.Hooks(f(g(h())))`.
func (c *InvaderClient) Use(hooks ...Hook) {
	c.hooks.Invader = append(c.hooks.Invader, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `invader.Intercept(f(g(h())))`.
func (c *InvaderClient) Intercept(interceptors ...Interceptor) {
	c.inters.Invader = append(c.inters.Invader, interceptors...)
}

// Create returns a builder for creating a Invader entity.
func (c *InvaderClient) Create() *InvaderCreate {
	mutation := newInvaderMutation(c.config, OpCreate)
	return &InvaderCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Invader entities.
func (c *InvaderClient) CreateBulk(builders ...*InvaderCreate) *InvaderCreateBulk {
	return &InvaderCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *InvaderClient) MapCreateBulk(slice any, setFunc func(*InvaderCreate, int)) *InvaderCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &InvaderCreateBulk{err: fmt.Errorf("calling to InvaderClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*InvaderCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &InvaderCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Invader.
func (c *InvaderClient) Update() *InvaderUpdate {
	mutation := newInvaderMutation(c.config, OpUpdate)
	return &InvaderUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *InvaderClient) UpdateOne(i *Invader) *InvaderUpdateOne {
	mutation := newInvaderMutation(c.config, OpUpdateOne, withInvader(i))
	return &InvaderUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *InvaderClient) UpdateOneID(id uuid.UUID) *InvaderUpdateOne {
	mutation := newInvaderMutation(c.config, OpUpdateOne, withInvaderID(id))
	return &InvaderUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Invader.
func (c *InvaderClient) Delete() *InvaderDelete {
	mutation := newInvaderMutation(c.config, OpDelete)
	return &InvaderDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *InvaderClient) DeleteOne(i *Invader) *InvaderDeleteOne {
	return c.DeleteOneID(i.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *InvaderClient) DeleteOneID(id uuid.UUID) *InvaderDeleteOne {
	builder := c.Delete().Where(invader.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &InvaderDeleteOne{builder}
}

// Query returns a query builder for Invader.
func (c *InvaderClient) Query() *InvaderQuery {
	return &InvaderQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeInvader},
		inters: c.Interceptors(),
	}
}

// Get returns a Invader entity by its id.
func (c *InvaderClient) Get(ctx context.Context, id uuid.UUID) (*Invader, error) {
	return c.Query().Where(invader.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *InvaderClient) GetX(ctx context.Context, id uuid.UUID) *Invader {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryEventID queries the event_id edge of a Invader.
func (c *InvaderClient) QueryEventID(i *Invader) *EventQuery {
	query := (&EventClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := i.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(invader.Table, invader.FieldID, id),
			sqlgraph.To(event.Table, event.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, invader.EventIDTable, invader.EventIDPrimaryKey...),
		)
		fromV = sqlgraph.Neighbors(i.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *InvaderClient) Hooks() []Hook {
	return c.hooks.Invader
}

// Interceptors returns the client interceptors.
func (c *InvaderClient) Interceptors() []Interceptor {
	return c.inters.Invader
}

func (c *InvaderClient) mutate(ctx context.Context, m *InvaderMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&InvaderCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&InvaderUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&InvaderUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&InvaderDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Invader mutation op: %q", m.Op())
	}
}

// MovementClient is a client for the Movement schema.
type MovementClient struct {
	config
}

// NewMovementClient returns a client for the Movement from the given config.
func NewMovementClient(c config) *MovementClient {
	return &MovementClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `movement.Hooks(f(g(h())))`.
func (c *MovementClient) Use(hooks ...Hook) {
	c.hooks.Movement = append(c.hooks.Movement, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `movement.Intercept(f(g(h())))`.
func (c *MovementClient) Intercept(interceptors ...Interceptor) {
	c.inters.Movement = append(c.inters.Movement, interceptors...)
}

// Create returns a builder for creating a Movement entity.
func (c *MovementClient) Create() *MovementCreate {
	mutation := newMovementMutation(c.config, OpCreate)
	return &MovementCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Movement entities.
func (c *MovementClient) CreateBulk(builders ...*MovementCreate) *MovementCreateBulk {
	return &MovementCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *MovementClient) MapCreateBulk(slice any, setFunc func(*MovementCreate, int)) *MovementCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &MovementCreateBulk{err: fmt.Errorf("calling to MovementClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*MovementCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &MovementCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Movement.
func (c *MovementClient) Update() *MovementUpdate {
	mutation := newMovementMutation(c.config, OpUpdate)
	return &MovementUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *MovementClient) UpdateOne(m *Movement) *MovementUpdateOne {
	mutation := newMovementMutation(c.config, OpUpdateOne, withMovement(m))
	return &MovementUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *MovementClient) UpdateOneID(id uuid.UUID) *MovementUpdateOne {
	mutation := newMovementMutation(c.config, OpUpdateOne, withMovementID(id))
	return &MovementUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Movement.
func (c *MovementClient) Delete() *MovementDelete {
	mutation := newMovementMutation(c.config, OpDelete)
	return &MovementDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *MovementClient) DeleteOne(m *Movement) *MovementDeleteOne {
	return c.DeleteOneID(m.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *MovementClient) DeleteOneID(id uuid.UUID) *MovementDeleteOne {
	builder := c.Delete().Where(movement.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &MovementDeleteOne{builder}
}

// Query returns a query builder for Movement.
func (c *MovementClient) Query() *MovementQuery {
	return &MovementQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeMovement},
		inters: c.Interceptors(),
	}
}

// Get returns a Movement entity by its id.
func (c *MovementClient) Get(ctx context.Context, id uuid.UUID) (*Movement, error) {
	return c.Query().Where(movement.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *MovementClient) GetX(ctx context.Context, id uuid.UUID) *Movement {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// QueryEventID queries the event_id edge of a Movement.
func (c *MovementClient) QueryEventID(m *Movement) *EventQuery {
	query := (&EventClient{config: c.config}).Query()
	query.path = func(context.Context) (fromV *sql.Selector, _ error) {
		id := m.ID
		step := sqlgraph.NewStep(
			sqlgraph.From(movement.Table, movement.FieldID, id),
			sqlgraph.To(event.Table, event.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, movement.EventIDTable, movement.EventIDPrimaryKey...),
		)
		fromV = sqlgraph.Neighbors(m.driver.Dialect(), step)
		return fromV, nil
	}
	return query
}

// Hooks returns the client hooks.
func (c *MovementClient) Hooks() []Hook {
	return c.hooks.Movement
}

// Interceptors returns the client interceptors.
func (c *MovementClient) Interceptors() []Interceptor {
	return c.inters.Movement
}

func (c *MovementClient) mutate(ctx context.Context, m *MovementMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&MovementCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&MovementUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&MovementUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&MovementDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Movement mutation op: %q", m.Op())
	}
}

// hooks and interceptors per client, for fast access.
type (
	hooks struct {
		Event, Invader, Movement []ent.Hook
	}
	inters struct {
		Event, Invader, Movement []ent.Interceptor
	}
)
