package models_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/models"
	"github.com/stretchr/testify/assert"
)

func TestCursor(t *testing.T) {
	// opaque pointer to a value of type T
	id := uuid.New()

	cursor := models.UUIDToCursor(id)
	idF, err := models.CursorToUUID(cursor)

	assert.NoError(t, err)
	assert.Equal(t, id, idF)
}

type mockEvent struct {
	eventID uuid.UUID
}

func (m mockEvent) GetEventID() uuid.UUID {
	return m.eventID
}
func (m mockEvent) GetDeviceID() string {
	return "my-device"
}
func (m mockEvent) GetEmittedAt() time.Time {
	return time.Now()
}
func (m mockEvent) GetType() models.EventType {
	return models.EventTypeMovement
}

func TestGeneratePaginationInfo(t *testing.T) {
	t.Parallel()

	events := make([]mockEvent, 0, 40)

	for i := 0; i < 40; i++ {
		events = append(events, mockEvent{eventID: uuid.Must(uuid.NewV7())})
	}

	t.Run("zero-event", func(t *testing.T) {
		t.Parallel()

		paginationInfo := models.GeneratePaginationInfo([]mockEvent{}, 15)

		assert.False(t, paginationInfo.HasNextPage)
		assert.Equal(t, "", paginationInfo.StartCursor)
		assert.Equal(t, "", paginationInfo.EndCursor)
	})

	t.Run("one-event-with-next-page", func(t *testing.T) {
		t.Parallel()

		paginationInfo := models.GeneratePaginationInfo(events[:2], 1)

		assert.True(t, paginationInfo.HasNextPage)
		assert.Equal(t, models.UUIDToCursor(events[0].GetEventID()), paginationInfo.StartCursor)
		assert.Equal(t, models.UUIDToCursor(events[0].GetEventID()), paginationInfo.EndCursor)
	})

	t.Run("one-event-without-next-page", func(t *testing.T) {
		t.Parallel()

		paginationInfo := models.GeneratePaginationInfo(events[:1], 1)

		assert.False(t, paginationInfo.HasNextPage)
		assert.Equal(t, models.UUIDToCursor(events[0].GetEventID()), paginationInfo.EndCursor)
		assert.Equal(t, models.UUIDToCursor(events[0].GetEventID()), paginationInfo.StartCursor)
	})

	t.Run("two-events-with-next-page", func(t *testing.T) {
		t.Parallel()

		paginationInfo := models.GeneratePaginationInfo(events[:3], 2)

		assert.True(t, paginationInfo.HasNextPage)
		assert.Equal(t, models.UUIDToCursor(events[0].GetEventID()), paginationInfo.StartCursor)
		assert.Equal(t, models.UUIDToCursor(events[1].GetEventID()), paginationInfo.EndCursor)
	})

	t.Run("two-events-without-next-page", func(t *testing.T) {
		t.Parallel()

		paginationInfo := models.GeneratePaginationInfo(events[:2], 2)

		assert.False(t, paginationInfo.HasNextPage)
		assert.Equal(t, models.UUIDToCursor(events[0].GetEventID()), paginationInfo.StartCursor)
		assert.Equal(t, models.UUIDToCursor(events[1].GetEventID()), paginationInfo.EndCursor)
	})

	t.Run("many-events-with-next-page", func(t *testing.T) {
		t.Parallel()

		paginationInfo := models.GeneratePaginationInfo(events, 20)

		assert.True(t, paginationInfo.HasNextPage)
		assert.Equal(t, models.UUIDToCursor(events[0].GetEventID()), paginationInfo.StartCursor)
		assert.Equal(t, models.UUIDToCursor(events[19].GetEventID()), paginationInfo.EndCursor)
	})

	t.Run("many-events-without-next-page", func(t *testing.T) {
		t.Parallel()

		paginationInfo := models.GeneratePaginationInfo(events, 40)

		assert.False(t, paginationInfo.HasNextPage)
		assert.Equal(t, models.UUIDToCursor(events[0].GetEventID()), paginationInfo.StartCursor)
		assert.Equal(t, models.UUIDToCursor(events[39].GetEventID()), paginationInfo.EndCursor)
	})
}
