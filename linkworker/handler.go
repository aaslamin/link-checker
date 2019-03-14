package linkworker

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	DB WorkerResultStorage
}

const (
	WorkersHandlerPath = "/workers"
)

func NewHandler(database WorkerResultStorage) *Handler {
	return &Handler{
		DB: database,
	}
}

func (h *Handler) SetupRoutes(r *httprouter.Router) {
	r.POST(WorkersHandlerPath, h.Create)
	r.GET(fmt.Sprintf("%s/%s", WorkersHandlerPath, ":id"), h.Get)
	r.DELETE(fmt.Sprintf("%s/%s", WorkersHandlerPath, ":id"), h.Remove)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var worker Worker
	if err := worker.ValidateWorkerRequest(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error creating new worker - error: %s", err)
		return
	}

	h.DB.AddWorker(r.Context(), worker.ID)
	go worker.ProcessURL(h.DB)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(worker)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	results, err := h.DB.GetWorkerResults(r.Context(), p.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	js, err := json.Marshal(results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := h.DB.RemoveWorker(r.Context(), p.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
