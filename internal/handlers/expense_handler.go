package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/bluesky-o/fairshare/internal/middleware"
	"github.com/bluesky-o/fairshare/internal/models"
	"github.com/bluesky-o/fairshare/internal/services"
)

type ExpenseHandler struct {
	expenseService *services.ExpenseService
}

func NewExpenseHandler(expenseService *services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	groupID, err := getGroupID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	var req models.CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	expense, err := h.expenseService.CreateExpense(r.Context(), userID, groupID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, expense)
}

func (h *ExpenseHandler) GetGroupExpenses(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	groupID, err := getGroupID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	expenses, err := h.expenseService.GetGroupExpenses(r.Context(), userID, groupID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, expenses)
}
