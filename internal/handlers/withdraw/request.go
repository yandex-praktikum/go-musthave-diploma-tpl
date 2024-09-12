package withdraw

type RequestBody struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}
