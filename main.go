package main

import (
	"bytes"
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
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
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

	go func() {
		// For pprof
		_ = http.ListenAndServe(":8081", nil)
	}()

	err := listenAndServe(":8080", make(chan string, 1))
	if err != nil {
		log.Println(err)
	}
}

func addHandlers() {
	addHandler_v1()
	addHandler_v2()
}

var (
	parserPool fastjson.ParserPool
)

func addHandler_v1() {
	http.HandleFunc("/v1", func(writer http.ResponseWriter, request *http.Request) {
		buf, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		var reqNames [][]byte

		parser := parserPool.Get()
		defer parserPool.Put(parser)

		js, err := parser.ParseBytes(buf)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		for _, name := range js.GetArray("names") {
			name := name.GetStringBytes()
			reqNames = append(reqNames, name)
		}

		sleepRaw := request.URL.Query().Get("sleep")
		if sleepRaw != "" {
			if sleep, _ := time.ParseDuration(sleepRaw); sleep > 0 {
				time.Sleep(sleep)
			}
		}

		hello2(reqNames, writer)
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

func fasthttpV1Handler(ctx *fasthttp.RequestCtx) {
	var reqNames [][]byte

	parser := parserPool.Get()
	defer parserPool.Put(parser)

	js, err := parser.ParseBytes(ctx.Request.Body())
	if err != nil {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		log.Println(err)
		return
	}

	for _, name := range js.GetArray("names") {
		name := name.GetStringBytes()
		reqNames = append(reqNames, name)
	}

	sleepRaw := ctx.QueryArgs().Peek("sleep")
	if len(sleepRaw) > 0 {
		if sleep, _ := time.ParseDuration(string(sleepRaw)); sleep > 0 {
			time.Sleep(sleep)
		}
	}

	hello2(reqNames, ctx.Response.BodyWriter())
}

func listenAndServe(listenAddr string, gotAddr chan<- string) error {
	server := &fasthttp.Server{
		Handler: fasthttpV1Handler,
	}
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
		_ = server.Shutdown()
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
		if strings.HasPrefix(name, "go") {
			_, _ = fmt.Fprintln(w, "Hello", name)
		}
	}
}

var (
	goStr = []byte("go")
)

func hello2(names [][]byte, w io.Writer) {
	for _, name := range names {
		if bytes.HasPrefix(name, goStr) {
			_, _ = fmt.Fprintf(w, "Hello %s\n", name)
		}
	}
}
