package main

import (
	"jobqueue/pool"
	"jobqueue/router"
	"jobqueue/service"
	"net/http"
)

func main() {
	pool := pool.New(4, nil)
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
