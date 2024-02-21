package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	resty "github.com/go-resty/resty/v2"
)

const (
	defaultLimit = 50
)

// ResponseWithHeight is a wrapper for returned values from REST API calls.
type ResponseWithHeight struct {
	Height string          `json:"height"`
	Result json.RawMessage `json:"result"`
}

// ReadRespWithHeight reads response with height that has been changed in REST APIs from v0.36.0.
func ReadRespWithHeight(resp *resty.Response) ResponseWithHeight {
	var responseWithHeight ResponseWithHeight
	err := json.Unmarshal(resp.Body(), &responseWithHeight)
	if err != nil {
		fmt.Printf("unmarshal responseWithHeight error - %s\n", err)
	}
	return responseWithHeight
}

// ParseHTTPArgs parses the request's URL and returns all arguments pairs.
// It separates page and limit used for pagination where a default limit can be provided.
func ParseHTTPArgs(r *http.Request) (from int64, limit int, err error) {
	fromStr := r.FormValue("from")
	if fromStr == "" {
		fromStr = "0"
	}

	from, err = strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		return from, limit, err
	}

	if from < 0 {
		return from, limit, fmt.Errorf("invalid value from : %d", from)
	}

	limitStr := r.FormValue("limit")
	if limitStr == "" {
		limitStr = "0"
	}

	limit, err = strconv.Atoi(limitStr) // ParseInt(limitStr, 10, 0)
	if err != nil {
		return from, limit, err
	}

	if limit <= 0 {
		return from, limit, fmt.Errorf("invalid value from : %d", limit)
	}

	if limit > defaultLimit {
		limit = defaultLimit
	}

	return from, limit, nil
}
