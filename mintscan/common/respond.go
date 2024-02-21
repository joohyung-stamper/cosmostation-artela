package common

import (
	"encoding/json"
	"net/http"
)

// respond responds result of any data type.
func respond(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	switch data := data.(type) {
	case []byte:
		// fmt.Printf("type : %T\n", data)
		w.Write(data)
	default:
		// fmt.Printf("type : %T\n", data)
		json.NewEncoder(w).Encode(data)
	}
}
