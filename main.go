package main

import (
	"jobqueue/router"
	"jobqueue/service"
	"jobqueue/worker"
	"net/http"
)

func main() {
	pool := worker.New(5, false)
	defer pool.Close()
	service := service.NewData(pool)
	r, err := router.New(service)
	if err != nil {
		panic(err)
	}
	if err := http.ListenAndServe("localhost:3000", r); err != nil {
		panic(err)
	}
}
