package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nixpig/syringe.sh/server/internal/user"
)

type HttpHandlers struct {
	userService user.UserService
}

func NewHttpHandlers(userService user.UserService) HttpHandlers {
	return HttpHandlers{userService}
}

func (h *HttpHandlers) PostUsersCreate(w http.ResponseWriter, r *http.Request) {
	var req user.RegisterUserRequestJsonDto

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("decode")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	createdUser, err := h.userService.Create(req)
	if err != nil {
		fmt.Println("create")
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(createdUser); err != nil {
		fmt.Println("encode")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
