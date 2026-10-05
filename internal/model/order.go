package model

// OrderRegistrationResult reports whether an order was inserted and, if not, who owns it.
type OrderRegistrationResult struct {
	Inserted bool
	Owner    string
}

// Order lifecycle statuses.
const (
	StatusRegistered = "REGISTERED"
	StatusInvalid    = "INVALID"
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
	StatusNew        = "NEW"
)
