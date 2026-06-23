package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bluesky-o/fairshare/internal/middleware"
	"github.com/bluesky-o/fairshare/internal/models"
	"github.com/bluesky-o/fairshare/internal/services"
	"github.com/go-chi/chi/v5"
)

type ExpenseHandler struct {
	expenseService *services.ExpenseService
}

func NewExpenseHandler(expenseService *services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

func getExpenseID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "expenseId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid expense id")
	}
	return id, nil
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

func (h *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	expenseID, err := getExpenseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	expense, err := h.expenseService.GetExpense(r.Context(), userID, expenseID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, expense)
}

func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	expenseID, err := getExpenseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid expense id")
		return
	}

	var req models.UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	expense, err := h.expenseService.UpdateExpense(r.Context(), userID, expenseID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, expense)
}
