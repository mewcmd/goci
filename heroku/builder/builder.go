package main

import (
	"code.google.com/p/gorilla/rpc"
	"code.google.com/p/gorilla/rpc/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

//env gets an environment variable with a default
func env(key, def string) (r string) {
	if r = os.Getenv(key); r == "" {
		r = def
	}
	return
}

//rpcServer is the rpc server for interacting with the builder
var rpcServer = rpc.NewServer()

func init() {
	//the rpcServer speaks jsonrpc
	rpcServer.RegisterCodec(json.NewCodec(), "application/json")
}

//bail is a helper function to run cleanup and panic
func bail(v interface{}) {
	cleanup.cleanup()
	if v != nil {
		panic(v)
	} else {
		os.Exit(0)
	}
}

func main() {
	//async run the setup and when that finishes announce
	go func() {
		if err := setup(); err != nil {
			bail(err)
		}
		if err := announce(); err != nil {
			bail(err)
		}

		//run the builder loop after a sucessful announce
		go builder()
	}()

	//set up the signal handler to bail and run cleanup
	signals := []os.Signal{
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGINT,
	}
	ch := make(chan os.Signal, len(signals))
	signal.Notify(ch, signals...)
	go func() {
		sig := <-ch
		log.Printf("Captured a %v\n", sig)
		bail(nil)
	}()

	http.Handle("/rpc", rpcServer)
	bail(http.ListenAndServe(":9080", nil))
}

//builder is a simple goroutine that 
func builder() {
	for {
		//get a task from the queue
		task := queue.pop()
		process(task)
	}
}
