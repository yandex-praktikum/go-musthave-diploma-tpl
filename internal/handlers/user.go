package handlers

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v4"
	"gophermart/internal/accrual"
	"gophermart/internal/middleware"
	"gophermart/internal/models"
	"gophermart/internal/repository"
	"gophermart/internal/service"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type UserHandler struct {
	UserService          service.UserService
	OrderService         service.OrderService
	WithdrawService      service.WithdrawService
	UserBalanceService   service.UserBalanceService
	AccrualSystemAddress string
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := uh.UserService.GetUserID(user.Username)

	if userID == repository.DatabaseError {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	if userID >= 0 {
		http.Error(w, "Failed to register user", http.StatusConflict)
		return
	}

	var err error
	tx, err := uh.OrderService.OrderRepository.DBStorage.Conn.BeginTx(
		uh.OrderService.OrderRepository.DBStorage.Ctx, pgx.TxOptions{})

	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	if user, err = uh.UserService.RegisterUser(user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		_ = tx.Rollback(uh.OrderService.OrderRepository.DBStorage.Ctx)
		return
	}

	if err = uh.UserBalanceService.CreateUserBalance(user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		_ = tx.Rollback(uh.OrderService.OrderRepository.DBStorage.Ctx)
		return
	}

	if err := tx.Commit(uh.OrderService.OrderRepository.DBStorage.Ctx); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	token, err := generateToken(user)

	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   3600,
	})

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds middleware.Credentials

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authUser, err := uh.UserService.AuthenticateUser(creds.Username, creds.Password)

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(authUser)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   3600,
	})

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func (uh *UserHandler) SaveOrder(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UsernameKey).(string)

	userID := uh.UserService.GetUserID(username)

	if userID < 0 {
		http.Error(w, "пользователь не найден", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Could not read body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	orderNumber := string(body)
	isDigit := isDigits(orderNumber)

	if !ValidateNumber(orderNumber) {
		http.Error(w, "Неверный формат номера заказа", http.StatusUnprocessableEntity)
		return
	}

	if !isDigit {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	result, err := uh.OrderService.IsOrderExist(orderNumber, userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if result == 1 {
		http.Error(w, "Номер заказа уже был загружен другим пользователем", http.StatusConflict)
		return
	} else if result == 2 {
		http.Error(w, "Номер заказа уже был загружен этим пользователем", http.StatusOK)
		return
	}

	err = uh.OrderService.SaveOrder(orderNumber, userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)

	go func() {
		tx, err := uh.OrderService.OrderRepository.DBStorage.Conn.BeginTx(
			uh.OrderService.OrderRepository.DBStorage.Ctx, pgx.TxOptions{})

		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		var registerResponse accrual.RegisterResponse
		registerResponse, err = accrual.GetOrderInfo(uh.AccrualSystemAddress, orderNumber)

		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		if registerResponse.Order != "" {
			err = uh.OrderService.UpdateOrder(orderNumber, registerResponse.Accrual, registerResponse.Status)

			if err != nil {
				_ = tx.Rollback(uh.OrderService.OrderRepository.DBStorage.Ctx)
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}

			err = uh.UserBalanceService.UpdateUserBalance(registerResponse.Accrual, userID)

			if err != nil {
				_ = tx.Rollback(uh.OrderService.OrderRepository.DBStorage.Ctx)
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
		}
	}()
}

func (uh *UserHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UsernameKey).(string)

	userID := uh.UserService.GetUserID(username)

	if userID < 0 {
		http.Error(w, "пользователь не найден", http.StatusNotFound)
		return
	}

	orderData, err := uh.OrderService.GetUserOrders(userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(orderData)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func (uh *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UsernameKey).(string)

	userID := uh.UserService.GetUserID(username)

	if userID < 0 {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	userBalance, err := uh.UserBalanceService.GetUserBalance(userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(userBalance)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func (uh *UserHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UsernameKey).(string)

	userID := uh.UserService.GetUserID(username)

	if userID < 0 {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	var withdraw models.Withdraw
	err := json.NewDecoder(r.Body).Decode(&withdraw)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isDigit := isDigits(withdraw.Order)

	if !isDigit {
		http.Error(w, "Неверный номер заказа", http.StatusUnprocessableEntity)
		return
	}

	result, err := uh.OrderService.IsOrderExist(withdraw.Order, userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if result == 1 {
		http.Error(w, "Номер заказа уже был загружен другим пользователем", http.StatusUnprocessableEntity)
		return
	}

	var code int
	code, err = uh.WithdrawService.Withdraw(userID, withdraw.Order, withdraw.Sum)

	if code == repository.WithdrawTransactionError {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if code == repository.NotEnoughFound {
		http.Error(w, "На счету недостаточно средств", http.StatusPaymentRequired)
		return
	}

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (uh *UserHandler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(middleware.UsernameKey).(string)

	userID := uh.UserService.GetUserID(username)

	if userID < 0 {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	withdrawalInfo, err := uh.WithdrawService.Withdrawals(userID)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(withdrawalInfo)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func ValidateNumber(orderNumber string) bool {
	orderNumber = strings.ReplaceAll(orderNumber, " ", "")
	orderNumber = strings.ReplaceAll(orderNumber, "-", "")

	for _, char := range orderNumber {
		if char < '0' || char > '9' {
			return false
		}
	}

	sum := 0
	alternate := false
	for i := len(orderNumber) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(orderNumber[i]))
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

func generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(middleware.SecretKey))
}

func isDigits(s string) bool {
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(s)
}
