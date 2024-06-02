package models_test

import (
	"testing"

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
