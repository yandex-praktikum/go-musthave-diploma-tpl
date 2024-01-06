package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"

	"github.com/akashipov/go-musthave-diploma-tpl/internal/server/general"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func ServerRouter(s *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/user/register", WithLogging(http.HandlerFunc(Register), s))
	r.Post("/api/user/login", WithLogging(http.HandlerFunc(Authontefication), s))
	r.Post("/api/user/orders", WithLogging(http.HandlerFunc(RequireAuth(LoadOrder)), s))
	r.Get("/api/user/orders", WithLogging(http.HandlerFunc(RequireAuth(GetOrders)), s))
	return r
}

func CheckContentTypeHeader(request *http.Request, value string) error {
	if request.Header.Get("Content-Type") != value {
		return errors.New("Wrong type on header Content-Type")
	}
	return nil
}

func Register(w http.ResponseWriter, request *http.Request) {
	err := CheckContentTypeHeader(request, "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	var buf bytes.Buffer
	var register RegisterData
	_, err = buf.ReadFrom(request.Body)
	defer request.Body.Close()
	err = json.Unmarshal(buf.Bytes(), &register)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Json data cannot be decoded"))
		return
	}
	encoder := hmac.New(sha256.New, []byte(*ServerKey))
	status, err := encoder.Write([]byte(register.Password))
	if err != nil {
		msg := fmt.Sprintf("Problem with hashing of password. Status is %d\n", status)
		status, err = w.Write([]byte(msg))
	} else {
		hashedPWD := encoder.Sum(nil)
		hashedPWDStr := base64.RawURLEncoding.EncodeToString(hashedPWD[:])
		checkUserQuery := "SELECT * FROM users WHERE login = $1"
		row := DB.QueryRowContext(
			request.Context(), checkUserQuery, register.Login,
		)
		var login string
		var pwd string
		err := row.Scan(&login, &pwd)
		if err != nil {
			// if there is no user, we can create new one
			if err.Error() == sql.ErrNoRows.Error() {
				createUserQuery := "INSERT INTO users VALUES($1, $2)"
				f := func() error {
					_, err := DB.ExecContext(
						request.Context(),
						createUserQuery, register.Login, hashedPWDStr,
					)
					return err
				}
				err = general.RetryCode(f, syscall.ECONNREFUSED)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					status, err = w.Write([]byte("Something is wrong, " + err.Error()))
				} else {
					request.Body = io.NopCloser(strings.NewReader(string(buf.Bytes())))
					request.ContentLength = int64(len(string(buf.Bytes())))
					w.WriteHeader(http.StatusOK)
					status, err = w.Write(
						[]byte("User was successfully created\nAuthorization...\n"),
					)
					Authontefication(w, request)
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				status, err = w.Write([]byte("Something is wrong, " + err.Error()))
			}
		} else {
			w.WriteHeader(http.StatusConflict)
			status, err = w.Write([]byte("User is already existed. Please set up other login"))
		}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Impossible to return answer. Status is: %v\n", status)
		return
	}
}

func DeleteTokenByUserIfExists(request *http.Request, name string) error {
	createUserQuery := "DELETE FROM tokens WHERE name = $1"
	f := func() error {
		_, err := DB.ExecContext(
			request.Context(),
			createUserQuery, name,
		)
		return err
	}
	err := general.RetryCode(f, syscall.ECONNREFUSED)
	return err
}

func WriteTokenToDB(request *http.Request, token string, exp int64, name string) error {
	createUserQuery := "INSERT INTO tokens VALUES($1, $2, TO_TIMESTAMP($3))"
	f := func() error {
		_, err := DB.ExecContext(
			request.Context(),
			createUserQuery, token, name, exp,
		)
		return err
	}
	err := general.RetryCode(f, syscall.ECONNREFUSED)
	return err
}

func Authontefication(w http.ResponseWriter, request *http.Request) {
	var register RegisterData
	var buf bytes.Buffer
	err := CheckContentTypeHeader(request, "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	_, err = buf.ReadFrom(request.Body)
	defer request.Body.Close()
	err = json.Unmarshal(buf.Bytes(), &register)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Json data cannot be decoded"))
		return
	}
	encoder := hmac.New(sha256.New, []byte(*ServerKey))
	status, err := encoder.Write([]byte(register.Password))
	if err != nil {
		msg := fmt.Sprintf("Problem with hashing of password. Status is %d\n", status)
		status, err = w.Write([]byte(msg))
	} else {
		hashedPWD := encoder.Sum(nil)
		hashedPWDStr := base64.RawURLEncoding.EncodeToString(hashedPWD[:])
		checkUserQuery := "SELECT * FROM users WHERE login = $1"
		row := DB.QueryRowContext(
			request.Context(), checkUserQuery, register.Login,
		)
		var login string
		var pwd string
		err := row.Scan(&login, &pwd)
		if err != nil {
			// if there is no user, we can create new one
			if err.Error() == sql.ErrNoRows.Error() {
				w.WriteHeader(http.StatusUnauthorized)
				status, err = w.Write([]byte("There is no user with this login"))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				status, err = w.Write([]byte("Something is wrong, " + err.Error()))
			}
		} else {
			if hashedPWDStr == pwd {
				exp := time.Now().Add(time.Hour * 24).Unix()
				claims := jwt.MapClaims{
					"username": register.Login,
					"exp":      exp,
				}
				token := jwt.NewWithClaims(
					jwt.SigningMethodHS256,
					claims,
				)
				tokenString, err := token.SignedString([]byte(*ServerKey))

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					status, err = w.Write([]byte(err.Error()))
				} else {
					// delete old version
					err := DeleteTokenByUserIfExists(request, register.Login)
					if err != nil {
						status, err = w.Write([]byte(err.Error()))
					} else {
						err := WriteTokenToDB(request, tokenString, exp, register.Login)
						if err != nil {
							status, err = w.Write([]byte(err.Error()))
						} else {
							w.Header().Set("Authorization", tokenString)
							w.WriteHeader(http.StatusAccepted)

							status, err = w.Write([]byte("User is successfully authorized."))
						}
					}
				}

			} else {
				w.WriteHeader(http.StatusUnauthorized)
				status, err = w.Write([]byte("Wrong password was passed."))
			}
		}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Impossible to return answer. Status is: %v\n", status)
		return
	}
}

func RequireAuth(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := CheckAuthoriazation(w, r, r.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		f(w, r)
	}
}

func CheckAuthoriazation(w http.ResponseWriter, request *http.Request, tkn string) (string, error) {
	query := "SELECT * FROM tokens WHERE token = $1"
	row := DB.QueryRowContext(
		request.Context(), query, tkn,
	)
	tkn = ""
	var name string
	var exp time.Time
	err := row.Scan(&tkn, &name, &exp)
	if err != nil {
		if err.Error() == sql.ErrNoRows.Error() {
			w.WriteHeader(http.StatusUnauthorized)
			return "", errors.New("Please login or register if you didn't do that")
		}
		return "", err
	}
	if time.Since(exp) <= 0 {
		return name, nil
	} else {
		return "", errors.New("Token is expired. Need to login again...")
	}
}

func GetNameFromToken(tkn string) (string, error) {
	hmacSecret := []byte(*ServerKey)
	token, err := jwt.Parse(
		tkn, func(token *jwt.Token) (interface{}, error) {
			return hmacSecret, nil
		},
	)
	if err != nil {
		return "", errors.New("Problem with parsing of JWT token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["username"].(string), nil
	} else {
		return "", errors.New("Invalid JWT Token")
	}
}

type Statuses int64

const (
	NEW Statuses = iota
	PROCESSING
	INVALID
	PROCESSED
)

func (s Statuses) String() string {
	switch s {
	case NEW:
		return "NEW"
	case PROCESSING:
		return "PROCESSING"
	case INVALID:
		return "INVALID"
	case PROCESSED:
		return "PROCESSED"
	}
	return "UNKNOWN"
}

func GetOrders(w http.ResponseWriter, request *http.Request) {
	name, err := GetNameFromToken(request.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad token value"))
		return
	}
	query := "SELECT * FROM orders WHERE name = $1 ORDER BY created_time"
	var orders Orders
	orders.Values = make([]OrderInfo, 0)
	f := func() error {
		rows, err := DB.QueryContext(
			request.Context(),
			query, name,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		var order OrderInfo
		for rows.Next() {
			err := rows.Scan(&order.OrderID, &order.Name, &order.CreatedTime)
			if err != nil {
				return err
			}
			client := resty.New()
			resp, err := client.R().Get(*ASAdress + "/api/orders/" + fmt.Sprint(order.OrderID))
			if err != nil {
				return err
			}
			if resp.StatusCode()%100 != 2 {
				return errors.New(string(resp.Body()))
			}
			err = json.Unmarshal(resp.Body(), &order)
			if err != nil {
				return err
			}
			fmt.Println(order.OrderID, order.Accrual, order.Status)
			orders.Values = append(orders.Values, order)
		}
		return rows.Err()
	}
	err = general.RetryCode(f, syscall.ECONNREFUSED)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if len(orders.Values) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		data, err := json.Marshal(orders)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
}

func LoadOrder(w http.ResponseWriter, request *http.Request) {
	err := CheckContentTypeHeader(request, "text/plain")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	name, err := GetNameFromToken(request.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad token value"))
		return
	}
	var buf bytes.Buffer
	defer request.Body.Close()
	_, err = buf.ReadFrom(request.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Problem with reading body"))
		return
	}
	order_id_str := string(buf.Bytes())
	if !LuhnChecksum(order_id_str) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("Luhn check is not passed"))
		return
	}
	order_id, err := strconv.Atoi(order_id_str)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
	}

	createOrderQuery := "INSERT INTO orders VALUES($1, $2, TO_TIMESTAMP($3))"
	f := func() error {
		_, err := DB.ExecContext(
			request.Context(),
			createOrderQuery, order_id, name, time.Now().Unix(),
		)
		return err
	}
	err = general.RetryCode(f, syscall.ECONNREFUSED)
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok {
			if pgErr.Code == "23505" {
				f := func() error {
					getOrder := "SELECT * FROM orders WHERE order_id = $1"
					row := DB.QueryRowContext(
						request.Context(),
						getOrder, order_id,
					)
					var n string
					var odID string
					var t time.Time
					err := row.Scan(&odID, &n, &t)
					if err != nil {
						return err
					}
					if n == name {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("This order was already processed by this user"))
					} else {
						w.WriteHeader(http.StatusConflict)
						w.Write([]byte("This order was already processed by other user"))
					}
					return nil
				}
				err = general.RetryCode(f, syscall.ECONNREFUSED)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
				}
				return
			}
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("OrderID has been taken..."))
}

func LuhnChecksum(input string) bool {
	var sum float64
	var value int
	if len(input)%2 == 0 {
		value = 0
	} else {
		value = 1
	}
	for i := 0; i < len(input)-1; i++ {
		digit, err := strconv.Atoi(string(input[i]))
		if err != nil {
			return false
		}
		if i%2 == value {
			sum += math.Mod(float64(digit)*2, 9)
		} else {
			sum += float64(digit)
		}
	}
	digit, err := strconv.Atoi(string(input[len(input)-1]))
	if err != nil {
		return false
	}
	return math.Mod(sum+float64(digit), 10) == 0
}
