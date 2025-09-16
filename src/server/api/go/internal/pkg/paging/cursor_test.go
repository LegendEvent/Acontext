package paging

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		id   uuid.UUID
	}{
		{
			name: "normal time and UUID",
			time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			id:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		},
		{
			name: "zero time",
			time: time.Time{},
			id:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		},
		{
			name: "current time",
			time: time.Now().UTC(),
			id:   uuid.New(),
		},
		{
			name: "Nil UUID",
			time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			id:   uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := EncodeCursor(tt.time, tt.id)

			// Verify cursor is not empty
			assert.NotEmpty(t, cursor)

			// Verify can decode back to original values
			decodedTime, decodedID, err := DecodeCursor(cursor)
			assert.NoError(t, err)
			assert.Equal(t, tt.time.UTC().UnixNano(), decodedTime.UnixNano())
			assert.Equal(t, tt.id, decodedID)
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	// First create some valid cursors for testing
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 123456789, time.UTC)
	testID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	validCursor := EncodeCursor(testTime, testID)

	tests := []struct {
		name    string
		cursor  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid cursor",
			cursor:  validCursor,
			wantErr: false,
		},
		{
			name:    "empty cursor",
			cursor:  "",
			wantErr: true,
			errMsg:  "empty cursor",
		},
		{
			name:    "invalid base64 encoding",
			cursor:  "invalid-base64!@#",
			wantErr: true,
		},
		{
			name:    "malformed cursor (missing separator)",
			cursor:  "MTcwNDE3NjQwMDAwMDAwMDAwMHNvbWV0aGluZw", // base64 encoded invalid format
			wantErr: true,
			errMsg:  "bad cursor",
		},
		{
			name:    "invalid timestamp",
			cursor:  "aW52YWxpZF90aW1lc3RhbXB8MTIzZTQ1NjctZTg5Yi0xMmQzLWE0NTYtNDI2NjE0MTc0MDAw", // "invalid_timestamp|123e4567-e89b-12d3-a456-426614174000"
			wantErr: true,
		},
		{
			name:    "invalid UUID",
			cursor:  "MTcwNDE3NjQwMDAwMDAwMDAwMHxpbnZhbGlkLXV1aWQ", // "1704176400000000000|invalid-uuid"
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decodedTime, decodedID, err := DecodeCursor(tt.cursor)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Equal(t, time.Time{}, decodedTime)
				assert.Equal(t, uuid.Nil, decodedID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, time.Time{}, decodedTime)
				assert.NotEqual(t, uuid.Nil, decodedID)
			}
		})
	}
}

func TestEncodeDecode_Roundtrip(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		id   uuid.UUID
	}{
		{
			name: "round trip test 1",
			time: time.Date(2024, 3, 15, 10, 30, 45, 123456789, time.UTC),
			id:   uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479"),
		},
		{
			name: "round trip test 2",
			time: time.Date(2023, 12, 31, 23, 59, 59, 999999999, time.UTC),
			id:   uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		},
		{
			name: "round trip test 3",
			time: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			id:   uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			cursor := EncodeCursor(tt.time, tt.id)
			assert.NotEmpty(t, cursor)

			// Decode
			decodedTime, decodedID, err := DecodeCursor(cursor)
			assert.NoError(t, err)

			// Verify round trip consistency
			assert.Equal(t, tt.time.UTC().UnixNano(), decodedTime.UnixNano())
			assert.Equal(t, tt.id, decodedID)
		})
	}
}

func TestCursor_EdgeCases(t *testing.T) {
	t.Run("very large timestamp", func(t *testing.T) {
		// Test cases near timestamp upper limit
		farFuture := time.Date(2099, 12, 31, 23, 59, 59, 999999999, time.UTC)
		testID := uuid.New()

		cursor := EncodeCursor(farFuture, testID)
		decodedTime, decodedID, err := DecodeCursor(cursor)

		assert.NoError(t, err)
		assert.Equal(t, farFuture.UnixNano(), decodedTime.UnixNano())
		assert.Equal(t, testID, decodedID)
	})

	t.Run("very small timestamp", func(t *testing.T) {
		// Test cases near timestamp lower limit
		earlyTime := time.Date(1970, 1, 1, 0, 0, 0, 1, time.UTC)
		testID := uuid.New()

		cursor := EncodeCursor(earlyTime, testID)
		decodedTime, decodedID, err := DecodeCursor(cursor)

		assert.NoError(t, err)
		assert.Equal(t, earlyTime.UnixNano(), decodedTime.UnixNano())
		assert.Equal(t, testID, decodedID)
	})

	t.Run("timezone handling", func(t *testing.T) {
		// Test if times in different timezones are correctly converted to UTC
		localTime := time.Date(2024, 6, 15, 14, 30, 0, 0, time.FixedZone("CST", 8*3600)) // UTC+8
		testID := uuid.New()

		cursor := EncodeCursor(localTime, testID)
		decodedTime, decodedID, err := DecodeCursor(cursor)

		assert.NoError(t, err)
		assert.Equal(t, localTime.UTC().UnixNano(), decodedTime.UnixNano())
		assert.Equal(t, testID, decodedID)
		assert.Equal(t, time.UTC, decodedTime.Location()) // Ensure decoded time is in UTC timezone
	})
}

func TestCursor_Consistency(t *testing.T) {
	t.Run("same input produces same cursor", func(t *testing.T) {
		testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		testID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

		cursor1 := EncodeCursor(testTime, testID)
		cursor2 := EncodeCursor(testTime, testID)

		assert.Equal(t, cursor1, cursor2)
	})

	t.Run("different inputs produce different cursors", func(t *testing.T) {
		testTime1 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		testTime2 := time.Date(2024, 1, 1, 12, 0, 0, 1, time.UTC) // 1 nanosecond difference
		testID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

		cursor1 := EncodeCursor(testTime1, testID)
		cursor2 := EncodeCursor(testTime2, testID)

		assert.NotEqual(t, cursor1, cursor2)
	})
}

func TestCursor_URLSafe(t *testing.T) {
	t.Run("cursor is URL safe", func(t *testing.T) {
		testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		testID := uuid.New()

		cursor := EncodeCursor(testTime, testID)

		// URL-safe base64 encoding should not contain these characters
		assert.NotContains(t, cursor, "+")
		assert.NotContains(t, cursor, "/")
		assert.NotContains(t, cursor, "=") // RawURLEncoding does not include padding characters
	})
}
