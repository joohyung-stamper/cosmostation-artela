package types

// KeyBase defines the structure of returned response from KeyBase API
type KeyBase struct {
	Status struct {
		Code int64  `json:"code"`
		Name string `json:"name"`
	} `json:"status"`
	Them []struct {
		ID       string `json:"id"`
		Pictures struct {
			Primary struct {
				URL    string `json:"url"`
				Source string `json:"source"`
			} `json:"primary"`
		} `json:"pictures"`
	} `json:"them"`
}
