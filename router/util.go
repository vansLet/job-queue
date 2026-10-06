package router

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"jobqueue/db"
	"jobqueue/model"
	"jobqueue/pool"
	"jobqueue/web"
	"net/http"
)

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
		fmt.Println(err)
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

func responeIdNotFound(w http.ResponseWriter, id int) {
	ErrRespone(w, http.StatusNotFound, map[string]any{
		"id":  id,
		"msg": db.ErrTaskNotFound.Error(),
	})
}

func handleErrPool(w http.ResponseWriter, err error) {

	if err != nil {
		e := err
		if errors.Is(e, pool.ErrChannelClose) {
			ErrRespone(w, http.StatusInternalServerError, "server closed")
			return
		} else if errors.Is(e, db.ErrCanceldDB) {
			ErrRespone(w, http.StatusOK, "cancle  success")
			return
		} else if errors.Is(e, db.ErrDBTimeOut) {
			ErrRespone(w, http.StatusRequestTimeout, "server to slow for proses")
			return
		} else {
			ErrRespone(w, http.StatusInternalServerError, "internal server error")
			return
		}

	}
}
