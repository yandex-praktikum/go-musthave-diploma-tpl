package order

type ResponseBody struct {
	Processing bool   `json:"processing"`
	Order      string `json:"order"`
}
