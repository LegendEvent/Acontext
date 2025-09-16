package serializer

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErr(t *testing.T) {
	tests := []struct {
		name    string
		errCode int
		msg     string
		err     error
		ginMode string
		wantErr bool
	}{
		{
			name:    "basic error response",
			errCode: http.StatusBadRequest,
			msg:     "test error",
			err:     nil,
			ginMode: gin.ReleaseMode,
			wantErr: false,
		},
		{
			name:    "response with error details (debug mode)",
			errCode: http.StatusInternalServerError,
			msg:     "server error",
			err:     errors.New("detailed error information"),
			ginMode: gin.DebugMode,
			wantErr: true,
		},
		{
			name:    "with error but release mode (no error details)",
			errCode: http.StatusInternalServerError,
			msg:     "server error",
			err:     errors.New("detailed error information"),
			ginMode: gin.ReleaseMode,
			wantErr: false,
		},
		{
			name:    "test mode shows error details",
			errCode: http.StatusBadRequest,
			msg:     "parameter error",
			err:     errors.New("parameter validation failed"),
			ginMode: gin.TestMode,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set Gin mode
			gin.SetMode(tt.ginMode)

			response := Err(tt.errCode, tt.msg, tt.err)

			assert.Equal(t, tt.errCode, response.Code)
			assert.Equal(t, tt.msg, response.Msg)
			assert.Nil(t, response.Data)

			if tt.wantErr {
				assert.NotEmpty(t, response.Error)
				if tt.err != nil {
					assert.Contains(t, response.Error, tt.err.Error())
				}
			} else {
				assert.Empty(t, response.Error)
			}
		})
	}

	// Reset to test mode
	gin.SetMode(gin.TestMode)
}

func TestDBErr(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		err     error
		wantMsg string
	}{
		{
			name:    "custom database error message",
			msg:     "user creation failed",
			err:     errors.New("duplicate key value"),
			wantMsg: "user creation failed",
		},
		{
			name:    "default database error message",
			msg:     "",
			err:     errors.New("connection timeout"),
			wantMsg: "database error",
		},
		{
			name:    "database error without error object",
			msg:     "data query failed",
			err:     nil,
			wantMsg: "data query failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set to debug mode to show error details
			gin.SetMode(gin.DebugMode)

			response := DBErr(tt.msg, tt.err)

			assert.Equal(t, http.StatusInternalServerError, response.Code)
			assert.Equal(t, tt.wantMsg, response.Msg)
			assert.Nil(t, response.Data)

			if tt.err != nil {
				assert.NotEmpty(t, response.Error)
				assert.Contains(t, response.Error, tt.err.Error())
			}
		})
	}

	// Reset to test mode
	gin.SetMode(gin.TestMode)
}

func TestParamErr(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		err     error
		wantMsg string
	}{
		{
			name:    "custom parameter error message",
			msg:     "user ID format error",
			err:     errors.New("invalid UUID format"),
			wantMsg: "user ID format error",
		},
		{
			name:    "default parameter error message",
			msg:     "",
			err:     errors.New("validation failed"),
			wantMsg: "parameter error",
		},
		{
			name:    "parameter error without error object",
			msg:     "missing required field",
			err:     nil,
			wantMsg: "missing required field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set to debug mode to show error details
			gin.SetMode(gin.DebugMode)

			response := ParamErr(tt.msg, tt.err)

			assert.Equal(t, http.StatusBadRequest, response.Code)
			assert.Equal(t, tt.wantMsg, response.Msg)
			assert.Nil(t, response.Data)

			if tt.err != nil {
				assert.NotEmpty(t, response.Error)
				assert.Contains(t, response.Error, tt.err.Error())
			}
		})
	}

	// Reset to test mode
	gin.SetMode(gin.TestMode)
}

func TestAuthErr(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		wantMsg string
	}{
		{
			name:    "custom authentication error message",
			msg:     "token expired",
			wantMsg: "token expired",
		},
		{
			name:    "default authentication error message",
			msg:     "",
			wantMsg: "authentication error",
		},
		{
			name:    "insufficient permission error",
			msg:     "insufficient permission",
			wantMsg: "insufficient permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := AuthErr(tt.msg)

			assert.Equal(t, http.StatusUnauthorized, response.Code)
			assert.Equal(t, tt.wantMsg, response.Msg)
			assert.Nil(t, response.Data)
			assert.Empty(t, response.Error) // AuthErr does not pass error object
		})
	}
}

func TestResponse_Structure(t *testing.T) {
	t.Run("verify Response structure", func(t *testing.T) {
		response := Response{
			Code:  http.StatusOK,
			Data:  map[string]string{"key": "value"},
			Msg:   "success",
			Error: "",
		}

		assert.Equal(t, http.StatusOK, response.Code)
		assert.NotNil(t, response.Data)
		assert.Equal(t, "success", response.Msg)
		assert.Empty(t, response.Error)
	})
}

func TestTrackedErrorResponse_Structure(t *testing.T) {
	t.Run("verify TrackedErrorResponse structure", func(t *testing.T) {
		response := TrackedErrorResponse{
			Response: Response{
				Code:  http.StatusInternalServerError,
				Msg:   "server error",
				Error: "detailed error information",
			},
			TraceID: "trace-123456",
		}

		assert.Equal(t, http.StatusInternalServerError, response.Code)
		assert.Equal(t, "server error", response.Msg)
		assert.Equal(t, "detailed error information", response.Error)
		assert.Equal(t, "trace-123456", response.TraceID)
	})
}

func TestErrorResponseInDifferentModes(t *testing.T) {
	testErr := errors.New("test error")

	modes := []struct {
		mode       string
		shouldShow bool
	}{
		{gin.ReleaseMode, false},
		{gin.DebugMode, true},
		{gin.TestMode, true},
	}

	for _, mode := range modes {
		t.Run("mode_"+mode.mode, func(t *testing.T) {
			gin.SetMode(mode.mode)

			response := Err(http.StatusBadRequest, "test message", testErr)

			if mode.shouldShow {
				assert.NotEmpty(t, response.Error)
				assert.Contains(t, response.Error, "test error")
			} else {
				assert.Empty(t, response.Error)
			}
		})
	}

	// Reset to test mode
	gin.SetMode(gin.TestMode)
}
