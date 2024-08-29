package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/eac0de/gophermart/internal/services"
	"github.com/eac0de/gophermart/pkg/middlewares"
)

type UserHandlers struct {
	userService *services.UserService
}

func NewUserHandlers(userService *services.UserService) *UserHandlers {
	return &UserHandlers{
		userService: userService,
	}
}

func (uh *UserHandlers) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromRequest(r)
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (uh *UserHandlers) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromRequest(r)
	defer r.Body.Close()
	err := uh.userService.DeleteUser(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (uh *UserHandlers) PatchUserHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromRequest(r)
	defer r.Body.Close()
	var requestBody struct {
		Name *string `json:"name"`
		Age  *uint8  `json:"age"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err := uh.userService.UpdateUser(r.Context(), user, requestBody.Name, requestBody.Age)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
