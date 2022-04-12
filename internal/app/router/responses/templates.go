package responses

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
)

const (
	// HEADERS ------------------------------------
	HeaderContentType     = "Content-Type"
	HeaderLocation        = "Location"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderXRealIP         = "X-Real-IP"
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderContentEncoding = "Content-Encoding"

	// CONTENT TYPE -------------------------------------------------------------
	charsetUTF8 = "charset=utf-8"

	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationXML             = "application/xml"
	MIMEApplicationXMLCharsetUTF8  = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                    = "text/xml"
	MIMETextXMLCharsetUTF8         = MIMETextXML + "; " + charsetUTF8
	MIMEApplicationForm            = "application/x-www-form-urlencoded"
	MIMETextPlain                  = "text/plain"
	MIMETextPlainCharsetUTF8       = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm              = "multipart/form-data"
)

// LoginData - структура данных логин/пароль пользователя
type LoginData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// BalanceData - структура для запроса на вывод данных о текущем состоянии бонусного счета пользователя
type BalanceData struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

// WithdrawData - структура входящего запроса на списание бонусов в счет оплаты заказа
type WithdrawData struct {
	Order string          `json:"order"`
	Sum   decimal.Decimal `json:"sum"`
}

// WithdrawalsData - структура для запроса на вывод данных обо всех операциях списания бонусов
type WithdrawalsData struct {
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
	ProcessedAt time.Time       `json:"processed_at"`
}

type OrderAccrualInfo struct {
	Order   string          `json:"order"`
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

func JSON(w http.ResponseWriter, code int, i interface{}) {
	data, err := json.Marshal(i)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	w.WriteHeader(code)
	w.Write(data)
}

func NoContent(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func WriteString(w http.ResponseWriter, code int, s string) {
	w.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	w.WriteHeader(code)
	w.Write([]byte(s)) // Проверять ошибку здесь смысла нет: если ты не можешь записать байтики в подключение, то ответ со статусом вернуть тоже не сможешь, тк соединение уже вероятно разорвано) (c)
}

// CreateCookie func provides a cookie "key=value" based on given params
func CreateCookie(key string, value string) *http.Cookie {
	return &http.Cookie{
		Name:    key,
		Value:   value,
		Path:    "/",
		Expires: time.Now().Add(time.Second * 60 * 60),
	}
}
