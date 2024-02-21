package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// [TODO]
// References
// 1. https://nordicapis.com/best-practices-api-error-handling/
// 2. https://stackoverflow.blog/2020/03/02/best-practices-for-rest-api-design/

// Facebook: https://developers.facebook.com/docs/graph-api/using-graph-api/error-handling/
// Twitter: https://developer.twitter.com/en/docs/basics/response-codes
// Gogle Maps Booking API: https://developers.google.com/maps-booking/reference/rest-api-v3/status_codes

// Organize error types needed for this project
// [1] Database
// 	FailedConnection

// [2] API Server
// 	RequestFailed
// 	OverMaxLimit
// 	RequiredParam
// 	InvalidParam
// 	Unauthorized

// [3] Database + API Server + General Format
// 	NotFound

// [3] General Format (Encoding / Decoding)
// 	InvalidFormat - address length, bech32 prefix

// 	FailedUnmarshalJSON
// 	FailedMarshalBinaryLengthPrefixed

// ErrorCode is the internal error code number.
type ErrorCode uint32

// ErrorMsg is the error message that is returns to client.
type ErrorMsg string

// WrapError parses the error into an object-like struct for exporting
type WrapError struct {
	ErrorCode ErrorCode `json:"error_code"`
	ErrorMsg  ErrorMsg  `json:"error_msg"`
}

const (
	InvalidFormat     ErrorCode = 202
	NotExist          ErrorCode = 203
	FailedConversion  ErrorCode = 204
	NotExistValidator ErrorCode = 205

	OverMaxLimit                      ErrorCode = 301
	FailedUnmarshalJSON               ErrorCode = 302
	FailedMarshalBinaryLengthPrefixed ErrorCode = 303

	NotFound ErrorCode = 404

	InternalServer    ErrorCode = 500
	ServerUnavailable ErrorCode = 501
	NoDataAvailable   ErrorCode = 502

	RequiredParam ErrorCode = 601
	InvalidParam  ErrorCode = 602
)

// ErrorCodeToErrorMsg returns error message from error code.
func ErrorCodeToErrorMsg(code ErrorCode) ErrorMsg {
	switch code {
	case InvalidFormat:
		return "Invalid format"
	case NotExist:
		return "NotExist"
	case NotExistValidator:
		return "NotExistValidator"
	case FailedConversion:
		return "FailedConversion"
	case OverMaxLimit:
		return "OverMaxLimit"
	case FailedUnmarshalJSON:
		return "FailedUnmarshalJSON"
	case FailedMarshalBinaryLengthPrefixed:
		return "FailedMarshalBinaryLengthPrefixed"
	case InternalServer:
		return "Internal server error"
	case ServerUnavailable:
		return "ServerUnavailable"
	case NoDataAvailable:
		return "NoDataAvailable"
	case NotFound:
		return "NotFound"
	default:
		return "Unknown"
	}
}

// ErrorCodeToErrorMsgs returns error message concatenating with custom message from error code.
func ErrorCodeToErrorMsgs(code ErrorCode, msg string) ErrorMsg {
	switch code {
	case RequiredParam:
		return ErrorMsg(msg)
	case InvalidParam:
		return ErrorMsg(msg)
	default:
		return "Unknown"
	}
}

// --------------------
// Error Types
// --------------------

func ErrInternalServer(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: InternalServer,
		ErrorMsg:  ErrorCodeToErrorMsg(InternalServer),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrInvalidFormat(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: InvalidFormat,
		ErrorMsg:  ErrorCodeToErrorMsg(InvalidFormat),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrNotExist(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: NotExist,
		ErrorMsg:  ErrorCodeToErrorMsg(NotExist),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrNotExistValidator(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: NotExistValidator,
		ErrorMsg:  ErrorCodeToErrorMsg(NotExistValidator),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrNotFound(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: NotFound,
		ErrorMsg:  ErrorCodeToErrorMsg(NotFound),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrFailedConversion(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: FailedConversion,
		ErrorMsg:  ErrorCodeToErrorMsg(FailedConversion),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrOverMaxLimit(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: OverMaxLimit,
		ErrorMsg:  ErrorCodeToErrorMsg(OverMaxLimit),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrFailedUnmarshalJSON(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: FailedUnmarshalJSON,
		ErrorMsg:  ErrorCodeToErrorMsg(FailedUnmarshalJSON),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrFailedMarshalBinaryLengthPrefixed(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: FailedMarshalBinaryLengthPrefixed,
		ErrorMsg:  ErrorCodeToErrorMsg(FailedMarshalBinaryLengthPrefixed),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrServerUnavailable(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: ServerUnavailable,
		ErrorMsg:  ErrorCodeToErrorMsg(ServerUnavailable),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrNoDataAvailable(w http.ResponseWriter, statusCode int) {
	wrapError := WrapError{
		ErrorCode: NoDataAvailable,
		ErrorMsg:  ErrorCodeToErrorMsg(ServerUnavailable),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrRequiredParam(w http.ResponseWriter, statusCode int, msg string) {
	wrapError := WrapError{
		ErrorCode: RequiredParam,
		ErrorMsg:  ErrorCodeToErrorMsgs(RequiredParam, msg),
	}
	PrintException(w, statusCode, wrapError)
}

func ErrInvalidParam(w http.ResponseWriter, statusCode int, msg string) {
	wrapError := WrapError{
		ErrorCode: InvalidParam,
		ErrorMsg:  ErrorCodeToErrorMsgs(InvalidParam, msg),
	}
	PrintException(w, statusCode, wrapError)
}

// --------------------
// PrintException
// --------------------

// PrintException prints out the exception result
func PrintException(w http.ResponseWriter, statusCode int, err WrapError) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode) // HTTP status code

	result, _ := json.Marshal(err)
	fmt.Fprint(w, string(result))
}
