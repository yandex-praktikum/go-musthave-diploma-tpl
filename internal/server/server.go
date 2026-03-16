package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"

	"gophermart/internal/accrual"
	"gophermart/internal/config"
	"gophermart/internal/migrations"
)

type Server struct {
	cfg           *config.Config
	db            *sql.DB
	mux           *http.ServeMux
	accrualClient *accrual.Client
}

func New(cfg *config.Config) (*Server, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := migrations.Apply(ctx, db); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}

	s := &Server{
		cfg: cfg,
		db:  db,
		mux: http.NewServeMux(),
	}

	if cfg.AccrualSystemAddr != "" {
		cl, err := accrual.New(cfg.AccrualSystemAddr)
		if err != nil {
			return nil, fmt.Errorf("create accrual client: %w", err)
		}
		s.accrualClient = cl
		go s.accrualWorker()
	}

	s.registerRoutes()

	return s, nil
}

func (s *Server) ListenAndServe() error {
	defer func() {
		if err := s.db.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}()

	return http.ListenAndServe(s.cfg.RunAddress, s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/user/register", s.handleRegister)
	s.mux.HandleFunc("/api/user/login", s.handleLogin)

	s.mux.HandleFunc("/api/user/orders", s.withAuth(s.handleOrders))
	s.mux.HandleFunc("/api/user/balance", s.withAuth(s.handleBalance))
	s.mux.HandleFunc("/api/user/balance/withdraw", s.withAuth(s.handleWithdraw))
	s.mux.HandleFunc("/api/user/withdrawals", s.withAuth(s.handleWithdrawals))
}

func (s *Server) accrualWorker() {
	if s.accrualClient == nil {
		return
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		rows, err := s.db.QueryContext(
			ctx,
			`SELECT id, number, user_id
			 FROM orders
			 WHERE status IN ('NEW', 'PROCESSING')
			 ORDER BY uploaded_at
			 LIMIT 100`,
		)
		cancel()
		if err != nil {
			log.Printf("accrualWorker: query orders: %v", err)
			time.Sleep(time.Second)
			continue
		}

		type row struct {
			id     int64
			number string
			userID int64
		}

		var batch []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.id, &r.number, &r.userID); err != nil {
				log.Printf("accrualWorker: scan row: %v", err)
				continue
			}
			batch = append(batch, r)
		}
		rows.Close()

		if len(batch) == 0 {
			time.Sleep(time.Second)
			continue
		}

		for _, ord := range batch {
			s.processAccrualOrder(ord.id, ord.number, ord.userID)
		}
	}
}

func (s *Server) processAccrualOrder(orderID int64, number string, userID int64) {
	if s.accrualClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := s.accrualClient.GetOrderInfo(ctx, number)
	if err != nil {
		if rl, ok := err.(*accrual.RateLimitError); ok {
			log.Printf("accrualWorker: rate limit reached, sleep %s", rl.RetryAfter)
			time.Sleep(rl.RetryAfter)
			return
		}
		log.Printf("accrualWorker: get order info: %v", err)
		return
	}

	if info == nil {
		return
	}

	switch info.Status {
	case accrual.StatusRegistered, accrual.StatusProcessing:
		if _, err := s.db.ExecContext(
			ctx,
			`UPDATE orders SET status = 'PROCESSING' WHERE id = $1`,
			orderID,
		); err != nil {
			log.Printf("accrualWorker: update order PROCESSING: %v", err)
		}
	case accrual.StatusInvalid:
		if _, err := s.db.ExecContext(
			ctx,
			`UPDATE orders SET status = 'INVALID' WHERE id = $1`,
			orderID,
		); err != nil {
			log.Printf("accrualWorker: update order INVALID: %v", err)
		}
	case accrual.StatusProcessed:
		var accrualVal float64
		if info.Accrual != nil {
			accrualVal = *info.Accrual
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			log.Printf("accrualWorker: begin tx: %v", err)
			return
		}
		defer tx.Rollback()

		if _, err := tx.ExecContext(
			ctx,
			`UPDATE orders
			 SET status = 'PROCESSED',
			     accrual = $1
			 WHERE id = $2`,
			accrualVal, orderID,
		); err != nil {
			log.Printf("accrualWorker: update order PROCESSED: %v", err)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("accrualWorker: commit tx: %v", err)
			return
		}
	}
}

func (s *Server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := s.currentUserID(r)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) currentUserID(r *http.Request) (int64, bool) {
	c, err := r.Cookie("user_id")
	if err != nil {
		return 0, false
	}
	id, err := strconv.ParseInt(c.Value, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var cred credentials
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if cred.Login == "" || cred.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var userID int64
	err = s.db.QueryRowContext(
		ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`,
		cred.Login, string(hash),
	).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	s.setUserCookie(w, userID)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var cred credentials
	if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if cred.Login == "" || cred.Password == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var (
		userID       int64
		passwordHash string
	)
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, password_hash FROM users WHERE login = $1`,
		cred.Login,
	).Scan(&userID, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(cred.Password)); err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	s.setUserCookie(w, userID)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) setUserCookie(w http.ResponseWriter, userID int64) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    strconv.FormatInt(userID, 10),
		Path:     "/",
		HttpOnly: true,
	})
}

func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateOrder(w, r)
	case http.MethodGet:
		s.handleListOrders(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (s *Server) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, _ := s.currentUserID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	number := strings.TrimSpace(string(body))
	if number == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !isValidOrderNumber(number) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var existingUserID int64
	err = s.db.QueryRowContext(
		ctx,
		`SELECT user_id FROM orders WHERE number = $1`,
		number,
	).Scan(&existingUserID)
	if err == nil {
		if existingUserID == userID {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4)`,
		userID, number, "NEW", time.Now(),
	)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleListOrders(w http.ResponseWriter, r *http.Request) {
	userID, _ := s.currentUserID(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT number, status, accrual, uploaded_at
		 FROM orders
		 WHERE user_id = $1
		 ORDER BY uploaded_at DESC`,
		userID,
	)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type orderResponse struct {
		Number     string   `json:"number"`
		Status     string   `json:"status"`
		Accrual    *float64 `json:"accrual,omitempty"`
		UploadedAt string   `json:"uploaded_at"`
	}

	var orders []orderResponse
	for rows.Next() {
		var (
			number     string
			status     string
			accrual    sql.NullFloat64
			uploadedAt time.Time
		)
		if err := rows.Scan(&number, &status, &accrual, &uploadedAt); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		var accrualPtr *float64
		if accrual.Valid {
			v := accrual.Float64
			accrualPtr = &v
		}
		orders = append(orders, orderResponse{
			Number:     number,
			Status:     status,
			Accrual:    accrualPtr,
			UploadedAt: uploadedAt.Format(time.RFC3339),
		})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

type balanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (s *Server) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, _ := s.currentUserID(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var accrued float64
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(accrual), 0)
		 FROM orders
		 WHERE user_id = $1 AND status = 'PROCESSED'`,
		userID,
	).Scan(&accrued); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var withdrawn float64
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(sum), 0)
		 FROM withdrawals
		 WHERE user_id = $1`,
		userID,
	).Scan(&withdrawn); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := balanceResponse{
		Current:   accrued - withdrawn,
		Withdrawn: withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

type withdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (s *Server) handleWithdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, _ := s.currentUserID(r)

	var req withdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if req.Order == "" || req.Sum <= 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !isValidOrderNumber(req.Order) {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var accrued float64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(accrual), 0)
		 FROM orders
		 WHERE user_id = $1 AND status = 'PROCESSED'`,
		userID,
	).Scan(&accrued); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var withdrawn float64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(sum), 0)
		 FROM withdrawals
		 WHERE user_id = $1`,
		userID,
	).Scan(&withdrawn); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	current := accrued - withdrawn
	if current < req.Sum {
		http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
		return
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO withdrawals (user_id, order, sum, processed_at)
		 VALUES ($1, $2, $3, $4)`,
		userID, req.Order, req.Sum, time.Now(),
	); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, _ := s.currentUserID(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT order, sum, processed_at
		 FROM withdrawals
		 WHERE user_id = $1
		 ORDER BY processed_at DESC`,
		userID,
	)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type withdrawalResponse struct {
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}

	var items []withdrawalResponse
	for rows.Next() {
		var (
			order       string
			sum         float64
			processedAt time.Time
		)
		if err := rows.Scan(&order, &sum, &processedAt); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		items = append(items, withdrawalResponse{
			Order:       order,
			Sum:         sum,
			ProcessedAt: processedAt.Format(time.RFC3339),
		})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func isValidOrderNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	var sum int
	double := false
	for i := len(s) - 1; i >= 0; i-- {
		d := int(s[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	const duplicateKey = "duplicate key value violates unique constraint"
	return strings.Contains(err.Error(), duplicateKey)
}
