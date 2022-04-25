package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
)

var (
	payload = bytes.NewBufferString(`{"names":["foo","bar","gopher","golang"]}`)

	serverAddr string
)

func Benchmark_addHandler_v1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp, err := http.Post("http://"+serverAddr+"/v1", "application/x-www-form-urlencoded", bytes.NewReader(payload.Bytes()))
		if err != nil {
			b.Fatal("http.Post:", err)
		}

		if resp.StatusCode != 200 {
			buf, _ := ioutil.ReadAll(resp.Body)
			b.Fatalf("Got status %d from server with body %q", resp.StatusCode, buf)
		}

		resp.Body.Close()
	}
}

func Benchmark_addHandler_v2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp, err := http.Post("http://"+serverAddr+"/v2", "application/x-www-form-urlencoded", bytes.NewReader(payload.Bytes()))
		if err != nil {
			b.Fatal("http.Post:", err)
		}

		if resp.StatusCode != 200 {
			buf, _ := ioutil.ReadAll(resp.Body)
			b.Fatalf("Got status %d from server with body %q", resp.StatusCode, buf)
		}

		resp.Body.Close()
	}
}

func TestMain(m *testing.M) {
	addHandlers()

	addrCh := make(chan string, 1)

	go func() {
		if err := listenAndServe("", addrCh); err != nil {
			log.Println("Can't listenAndServe:", err)
			os.Exit(1)
		}
	}()

	var ok bool
	serverAddr, ok = <-addrCh
	if !ok {
		log.Println("Can't read from the addr chan")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
