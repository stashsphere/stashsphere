package resources

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/stashsphere/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestNotificationFromModelJSON(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name        string
		contentType string
		content     string
		expected    string
	}{
		{
			name:        "known content type",
			contentType: "THING_SHARED",
			content:     `{"thingId": "t1", "sharerId": "u1"}`,
			expected:    `{"id":"n1","createdAt":"2026-01-02T03:04:05Z","acknowledged":true,"contentType":"THING_SHARED","content":{"thingId":"t1","sharerId":"u1"}}`,
		},
		{
			name:        "unknown content type is passed through",
			contentType: "SOMETHING_NEW",
			content:     `{"foo": 1}`,
			expected:    `{"id":"n1","createdAt":"2026-01-02T03:04:05Z","acknowledged":true,"contentType":"SOMETHING_NEW","content":{"foo":1}}`,
		},
		{
			name:        "malformed content is passed through",
			contentType: "LIST_SHARED",
			content:     `{"listId": 42}`,
			expected:    `{"id":"n1","createdAt":"2026-01-02T03:04:05Z","acknowledged":true,"contentType":"LIST_SHARED","content":{"listId":42}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &models.Notification{
				ID:             "n1",
				ContentType:    tt.contentType,
				Content:        []byte(tt.content),
				CreatedAt:      createdAt,
				AcknowledgedAt: null.TimeFrom(createdAt),
			}
			out, err := json.Marshal(NotificationFromModel(model))
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(out))
		})
	}
}
