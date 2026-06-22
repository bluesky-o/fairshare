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

type GroupHandler struct {
	groupService *services.GroupService
}

func NewGroupHandler(groupService *services.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

func getGroupID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid group id")
	}
	return id, nil
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	group, err := h.groupService.CreateGroup(r.Context(), userID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, group)
}

func (h *GroupHandler) GetMyGroups(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	groups, err := h.groupService.GetMyGroups(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, groups)
}

func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	groupID, err := getGroupID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	group, err := h.groupService.GetGroup(r.Context(), userID, groupID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, group)
}

func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	groupID, err := getGroupID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	var req models.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	group, err := h.groupService.UpdateGroup(r.Context(), userID, groupID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, group)
}

func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	groupID, err := getGroupID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	if err := h.groupService.DeleteGroup(r.Context(), userID, groupID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, map[string]string{
		"message": "group deleted successfully",
	})
}

