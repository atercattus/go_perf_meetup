package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"regexp"
	"time"
)

func main() {
	// // runtime.MemProfileRate = 0
	//
	// defer func() {
	// 	fd, _ := os.Create("mem.pprof")
	// 	runtime.GC()
	// 	_ = pprof.WriteHeapProfile(fd)
	// 	_ = fd.Close()
	// }()
	//
	// fd, _ := os.Create("cpu.pprof")
	// _ = pprof.StartCPUProfile(fd)
	// defer func() {
	// 	pprof.StopCPUProfile()
	// 	_ = fd.Close()
	// }()

	addHandlers()

	err := listenAndServe(":8080", make(chan string, 1))
	if err != nil {
		log.Println(err)
	}
}

func addHandlers() {
	addHandler_v1()
	addHandler_v2()
}

func addHandler_v1() {
	http.HandleFunc("/v1", func(writer http.ResponseWriter, request *http.Request) {
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		var req struct {
			Names []string
		}

		if err := json.Unmarshal(buf, &req); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		sleepRaw := request.URL.Query().Get("sleep")
		if sleep, _ := time.ParseDuration(sleepRaw); sleep > 0 {
			time.Sleep(sleep)
		}

		hello(req.Names, writer)
	})
}

func addHandler_v2() {
	http.HandleFunc("/v2", func(writer http.ResponseWriter, request *http.Request) {
		var req struct {
			Names []string
		}

		if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		sleepRaw := request.URL.Query().Get("sleep")
		if sleep, _ := time.ParseDuration(sleepRaw); sleep > 0 {
			time.Sleep(sleep)
		}

		hello(req.Names, writer)
	})
}

func listenAndServe(listenAddr string, gotAddr chan<- string) error {
	server := &http.Server{}
	ln, err := net.Listen("tcp4", listenAddr)
	if err != nil {
		close(gotAddr)
		return fmt.Errorf("listen: %w", err)
	}

	gotAddr <- ln.Addr().String()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
		_ = server.Close()
	}()

	log.Println("Listening", ln.Addr())
	err = server.Serve(ln)
	if err != nil {
		_ = ln.Close()
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

func hello(names []string, w io.Writer) {
	for _, name := range names {
		if ok, _ := regexp.Match("^go", []byte(name)); ok {
			_, _ = fmt.Fprintln(w, "Hello", name)
		}
	}
}
