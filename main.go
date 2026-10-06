package main

import (
	"context"
	"errors"
	"fmt"
	"jobqueue/db"
	"jobqueue/pool"
	"jobqueue/router"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	pool := pool.New(5, nil)
	db, err := db.Init("data/tasks.db", pool)
	if err != nil {
		panic(err)
	}
	mux, err := router.New(db)
	if err != nil {
		panic(err)
	}
	server := http.Server{
		Addr:        ":3000",
		Handler:     mux,
		IdleTimeout: time.Second * 5,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			fmt.Println(err)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	server.Shutdown(ctx)
	db.Close()
	pool.Close()
}
