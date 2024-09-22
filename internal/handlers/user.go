package handlers

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/sub3er0/gophermart/internal/middleware"
	"github.com/sub3er0/gophermart/internal/models"
	"github.com/sub3er0/gophermart/internal/service"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type UserHandler struct {
	UserService     service.UserService
	OrderService    service.OrderService
	WithdrawService service.WithdrawService
}

type Withdraw struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}

func (uh *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := uh.UserService.UserRepository.IsUserExists(user.Username)

	if userID == -2 {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	if userID >= 0 {
		http.Error(w, "Failed to register user", http.StatusConflict)
		return
	}

	var err error

	if user, err = uh.UserService.RegisterUser(user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
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
	w.WriteHeader(http.StatusCreated)

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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(middleware.SecretKey))
}

func (uh *UserHandler) SaveOrder(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	userID := uh.UserService.UserRepository.IsUserExists(username)

	if userID < 0 {
		http.Error(w, "пользователь не найден", http.StatusNotFound)
		return
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Could not read body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	bodyString := string(body)
	isDigit := isDigits(bodyString)

	if isDigit != true {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	result, err := uh.OrderService.IsOrderExist(bodyString, userID)

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

	err = uh.OrderService.SaveOrder(bodyString, userID, 11)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func isDigits(s string) bool {
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(s)
}

func (uh *UserHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	userID := uh.UserService.UserRepository.IsUserExists(username)

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

	return
}

func (uh *UserHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	userID := uh.UserService.UserRepository.IsUserExists(username)

	if userID < 0 {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	userBalance, err := uh.OrderService.GetUserBalance(userID)

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

	return
}

func (uh *UserHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	userID := uh.UserService.UserRepository.IsUserExists(username)

	if userID < 0 {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	var withdraw Withdraw
	err := json.NewDecoder(r.Body).Decode(&withdraw)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	isDigit := isDigits(withdraw.Order)

	if isDigit != true {
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

	if code == -1 {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	if code == -2 {
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
	username := r.Context().Value("username").(string)

	userID := uh.UserService.UserRepository.IsUserExists(username)

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

	return
}
