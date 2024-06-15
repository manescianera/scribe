package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	log.Println("Creating new API handler...")
	return &Handler{db: db}
}

func (h *Handler) StartAPI() {
	log.Println("Starting API server...")
	router := mux.NewRouter()
	router.HandleFunc("/articles", h.ArticlesHandler).Methods("GET")
	router.HandleFunc("/articles/{id}", h.ArticleHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func (h *Handler) ArticlesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, file_name, transcription FROM transcriptions")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.FileName, &a.Transcription); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		articles = append(articles, a)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(articles); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ArticleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	var a Article
	err = h.db.QueryRow("SELECT id, file_name, transcription FROM transcriptions WHERE id = $1", id).Scan(&a.ID, &a.FileName, &a.Transcription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(a); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
