package types

type ResponseMeta struct {
	Errors []Error `json:"errors,omitempty"`
}

func (rm ResponseMeta) FromErr(field string, err error) ResponseMeta {
	rm.Errors = []Error{{Field: field, Message: err.Error()}}

	return rm
}

func (rm ResponseMeta) FromMessage(field, description string) ResponseMeta {
	rm.Errors = []Error{{Field: field, Message: description}}

	return rm
}

type Error struct {
	Field   string `binding:"required" json:"field"`
	Message string `binding:"required" json:"message"`
}
