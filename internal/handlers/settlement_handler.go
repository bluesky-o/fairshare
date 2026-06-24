package handlers

import (
	"net/http"

	"github.com/bluesky-o/fairshare/internal/middleware"
	"github.com/bluesky-o/fairshare/internal/services"
)

type SettlementHandler struct {
	balanceService *services.BalanceService
}

func NewSettlementHandler(balanceService *services.BalanceService) *SettlementHandler {
	return &SettlementHandler{
		balanceService: balanceService,
	}
}

func (h *SettlementHandler) GetGroupBalances(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	groupID, err := getGroupID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
	}

	balance, err := h.balanceService.GetGroupBalances(r.Context(), userID, groupID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, balance)
}
