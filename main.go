package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	upstream := os.Getenv("AS_PROXY_TO_ADDR")
	bind:= os.Getenv("AS_PROXY_BIND")

	log.Println("Preparing local server...")
	rtr := mux.NewRouter()
	rtr.HandleFunc("/_matrix/client/r0/register", func(w http.ResponseWriter, r *http.Request) {
		log.Println()
		defer dumpAndCloseStream(r.Body)

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		i := make(map[string]interface{})
		err = json.Unmarshal(b, &i)
		if err != nil {
			log.Fatal(err)
		}
		if _, ok := i["type"]; !ok {
			i["type"] = "m.login.application_service"
		}

		j, err := json.Marshal(i)
		if err != nil {
			log.Fatal(err)
		}

		r2, err := http.NewRequest(r.Method, upstream + r.RequestURI, bytes.NewBuffer(j))
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range r.Header {
			r2.Header.Set(k, v[0])
		}
		resp, err := http.DefaultClient.Do(r2)
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range resp.Header {
			w.Header().Set(k, v[0])
		}
		defer dumpAndCloseStream(resp.Body)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	})
	srv := &http.Server{Addr: bind, Handler: rtr}
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, os.Kill)
	go func() {
		defer close(stop)
		<-stop
		log.Println("Stopping local server...")
		_ = srv.Close()
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("Goodbye!")
}

func dumpAndCloseStream(r io.ReadCloser) {
	if r == nil {
		return // nothing to dump or close
	}
	_, _ = io.Copy(ioutil.Discard, r)
	_ = r.Close()
}