package serializer

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response
type Response struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data,omitempty"`
	Msg   string      `json:"msg"`
	Error string      `json:"error,omitempty"`
}

// TraceErrorResponse
type TrackedErrorResponse struct {
	Response
	TraceID string `json:"trace_id"`
}

// CheckLogin
func CheckLogin() Response {
	return Response{
		Code: http.StatusUnauthorized,
		Msg:  "please login first",
	}
}

// Err
func Err(errCode int, msg string, err error) Response {
	res := Response{
		Code: errCode,
		Msg:  msg,
	}
	// development mode, show error detail
	if err != nil && gin.Mode() != gin.ReleaseMode {
		res.Error = fmt.Sprintf("%+v", err)
	}
	return res
}

// DBErr
func DBErr(msg string, err error) Response {
	if msg == "" {
		msg = "database error"
	}
	return Err(http.StatusInternalServerError, msg, err)
}

// ParamErr
func ParamErr(msg string, err error) Response {
	if msg == "" {
		msg = "parameter error"
	}
	return Err(http.StatusBadRequest, msg, err)
}
