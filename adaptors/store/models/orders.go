package models

const (
	NEW = 1 << iota
	PROCESSING
	INVALID
	PROCESSED
)

type StatusOrder int

func (s StatusOrder) String() string {
	switch s {
	case NEW:
		return "NEW"
	case PROCESSING:
		return "PROCESSING"
	case INVALID:
		return "INVALID"
	case PROCESSED:
		return "PROCESSED"
	default:
		return "BROKEN"
	}
}

func StatusOrderToInt(s string) StatusOrder {
	switch s {
	case "NEW":
		return NEW
	case "PROCESSING":
		return PROCESSING
	case "INVALID":
		return INVALID
	case "PROCESSED":
		return PROCESSED
	default:
		return INVALID
	}
}

type AddOrderData struct {
	UserID  string
	OrderID string
	Status  StatusOrder
}

func NewAddOrderData(userID, orderID string) AddOrderData {
	return AddOrderData{
		UserID:  userID,
		OrderID: orderID,
		Status:  NEW,
	}
}

type GetOwnerForOrderData struct {
	OrderID string
}

type GetOrdersData struct {
	UserID string
}

type GetOrdersDataResult struct {
	Orders []OrderData
}

type OrderData struct {
	Number     string
	Status     StatusOrder
	Accrual    int
	UploadedAt string
}

func NewOrderData(orderID, status string, accrual int) OrderData {
	return OrderData{
		Number:  orderID,
		Status:  StatusOrderToInt(status),
		Accrual: accrual,
	}
}

type GetUserBalanceData struct {
	UserID string
}

type GetUserBalanceDataResult struct {
	Current   float64
	Withdrawn float64
}

type WithdrawData struct {
	UserID  string
	OrderID string
	Sum     float64
}

type WithdrawalsData struct {
	UserID string
}

type WithdrawsDataResult struct {
	Data []UserWithdraw
}

type UserWithdraw struct {
	OrderID     string
	Sun         float64
	ProcessedAt string
}

type LockNewOrders struct {
	Orders []string
}
