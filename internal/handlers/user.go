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

func (uh *UserHandlers) UserHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromRequest(r)
	defer r.Body.Close()
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
		return
	} else if r.Method == http.MethodDelete {
		err := uh.userService.DeleteUser(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	} else if r.Method == http.MethodPatch {
		var requestBody struct {
			Name  *string `json:"name"`
			Age   *uint8  `json:"age"`
			Email *string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		err := uh.userService.UpdateUser(r.Context(), user, requestBody.Name, requestBody.Age, requestBody.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}
