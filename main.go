package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
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

	_ = http.ListenAndServe(":8080", nil)
}

func hello(names []string, w io.Writer) {
	for _, name := range names {
		if ok, _ := regexp.Match("^go", []byte(name)); ok {
			_, _ = fmt.Fprintln(w, "Hello", name)
		}
	}
}
