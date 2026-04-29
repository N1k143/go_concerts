package handlers

import (
	"concerts/internal/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ConcertHandler struct {
	repo *repository.ConcertRepo
}

func NewConcertHandler(repo *repository.ConcertRepo) *ConcertHandler {
	return &ConcertHandler{repo: repo}
}

func (h *ConcertHandler) List(w http.ResponseWriter, r *http.Request) {
	concerts, err := h.repo.ListAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"concerts": concerts})
}

func (h *ConcertHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "concert-id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusNotFound, "A concert with this ID does not exist")
		return
	}

	concert, err := h.repo.GetByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if concert == nil {
		writeError(w, http.StatusNotFound, "A concert with this ID does not exist")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"concert": concert})
}
