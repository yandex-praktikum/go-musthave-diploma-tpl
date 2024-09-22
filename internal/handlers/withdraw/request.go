package withdraw

type RequestBody struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}
