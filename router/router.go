package router

import (
	"encoding/json/v2"
	"jobqueue/model"
	"jobqueue/service"
	"jobqueue/web"
	"net/http"
	"strconv"
)

func New(s *service.DataTask) (*http.ServeMux, error) {
	r := TaskRouter{data: s}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", r.CreateJob)
	mux.HandleFunc("GET /jobs", r.GetJobs)
	mux.HandleFunc("GET /jobs/{id}", r.GetById)
	mux.HandleFunc("DELETE /jobs/{id}", r.DeleteById)
	return mux, nil
}

func ErrRespone(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	d := map[string]any{
		"err": data,
	}
	if err := json.MarshalWrite(w, d); err != nil {
		responeServerErr(w, http.StatusInternalServerError, "internal server error")

	}
}

func SuccRespone(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	d := map[string]any{
		"data": data,
	}
	if err := json.MarshalWrite(w, d); err != nil {
		responeServerErr(w, http.StatusInternalServerError, "internal server error")
	}
}

func taskFromRequest(r *http.Request) (model.Task, error) {
	defer r.Body.Close()
	var newTask web.TaskRequest
	err := json.UnmarshalRead(r.Body, &newTask)
	if err != nil {
		return model.Task{}, err
	}
	return web.FromRequest(newTask), nil
}

func responeServerErr(w http.ResponseWriter, status int, err string) {
	http.Error(w, err, status)
}

type TaskRouter struct {
	data *service.DataTask
}

func (task *TaskRouter) CreateJob(w http.ResponseWriter, r *http.Request) {
	newTask, err := taskFromRequest(r)
	if err != nil {
		ErrRespone(w, http.StatusBadRequest, "format json invalid")
		return
	}

	s, err := task.data.CreateTask(newTask)
	if err != nil {
		ErrRespone(w, http.StatusBadRequest, err)
		return
	}
	SuccRespone(w, http.StatusCreated, web.ToRespone(s))
}

func (task *TaskRouter) GetJobs(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	s := task.data.GetAll()
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
	s, err := task.data.GetById(id)
	if err != nil {
		e := map[string]any{
			"id":  id,
			"msg": err.Error(),
		}
		ErrRespone(w, http.StatusNotFound, e)
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
	err = task.data.Delete(id)
	if err != nil {
		e := map[string]any{
			"id":  id,
			"msg": err.Error(),
		}
		ErrRespone(w, http.StatusNotFound, e)
		return
	}
	s := map[string]any{
		"id":  id,
		"msg": "Success Delete",
	}
	SuccRespone(w, http.StatusOK, s)
}
