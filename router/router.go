package router

import (
	"fmt"
	"jobqueue/db"
	"jobqueue/web"
	"net/http"
	"strconv"
)

func New(db db.DatabaseTasks) (*http.ServeMux, error) {
	r := TaskRouter{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", r.CreateJob)
	mux.HandleFunc("GET /jobs", r.GetJobs)
	mux.HandleFunc("GET /jobs/{id}", r.GetById)
	mux.HandleFunc("DELETE /jobs/{id}", r.DeleteById)
	return mux, nil
}

type TaskRouter struct {
	db db.DatabaseTasks
}

func (task *TaskRouter) CreateJob(w http.ResponseWriter, r *http.Request) {
	newTask, err := taskFromRequest(r)
	if err != nil {
		ErrRespone(w, http.StatusBadRequest, "format json invalid")
		return
	}

	s, err := task.db.Add(r.Context(), newTask)
	if err != nil {
		fmt.Println(err)
		handleErrPool(w, err)
		return
	}
	SuccRespone(w, http.StatusCreated, web.ToRespone(s))
}

func (task *TaskRouter) GetJobs(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	s := task.db.Gets(r.Context())
	d := make([]web.ResponeTask, 0, len(s))
	for _, t := range s {
		d = append(d, web.ToRespone(t))
	}
	SuccRespone(w, http.StatusOK, d)
}

func (task *TaskRouter) GetById(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	val := r.PathValue("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		ErrRespone(w, http.StatusBadRequest, "\"id\" not format number")
		return
	}
	s, err := task.db.GetById(r.Context(), id)
	if err != nil {
		responeIdNotFound(w, id)
		return
	}
	SuccRespone(w, http.StatusOK, web.ToRespone(s))
}

func (task *TaskRouter) DeleteById(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	val := r.PathValue("id")
	id, err := strconv.Atoi(val)
	if err != nil {
		ErrRespone(w, http.StatusBadRequest, "\"id\" not format number")
		return
	}
	err = task.db.Remove(r.Context(), id)
	if err != nil {
		responeIdNotFound(w, id)
		return
	}
	s := map[string]any{
		"id":  id,
		"msg": "Success Delete",
	}
	SuccRespone(w, http.StatusOK, s)
}
