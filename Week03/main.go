package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HttpServer struct {
	server  *http.Server
	handler http.Handler
	ctx     context.Context
}

var done = make(chan int)

func main() {

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {

		mux := http.NewServeMux()
		mux.HandleFunc("/close", func(writer http.ResponseWriter, request *http.Request) {
			done <- 1
		})

		hs := NewHttpServer(":8099", mux)
		go func() {
			err := hs.Start()
			if err != nil {
				panic(err)
			}
		}()

		select {
		case <-done:
			return hs.Stop()
		case <-ctx.Done():
			return errors.New("关闭")
		}
	})

	g.Go(func() error {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGEMT)
		select {
		case <-sig:
			return errors.New("关闭")
		case <-ctx.Done():
			return errors.New("关闭")
		}
	})

	err := g.Wait()
	fmt.Println(err)
}

func NewHttpServer (addr string, mux http.Handler) *HttpServer{
	hs := &HttpServer{}
	hs.server = &http.Server{
		Addr:addr,
		WriteTimeout:time.Second*5,
		Handler:mux,
	}
	return hs
}

func (hs *HttpServer) Start() error {
	fmt.Printf("httpServer start")
	return hs.server.ListenAndServe()
}

func (hs *HttpServer) Stop() error {
	err := hs.server.Shutdown(hs.ctx)
	return err
}
