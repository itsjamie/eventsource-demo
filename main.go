package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var msgChannel = make(chan string, 10)

func main() {
	fs := http.FileServer(http.Dir("public"))
	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", fs))
	mux.Handle("/message", http.HandlerFunc(postMessage))
	mux.Handle("/events", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "text/event-stream" {
			w.WriteHeader(http.StatusNotAcceptable)
			io.WriteString(w, "not an eventsource request")
			return
		}

		headers := w.Header()
		headers.Set("Content-Type", "text/event-stream; charset=utf-8")
		headers.Set("Cache-Control", "no-cache")
		headers.Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		writeMsg := func(t time.Time) {
			io.WriteString(w, fmt.Sprintf(
				`id: %d
data: {"msg":"We're saying hello at %s"}

`,
				t.Unix(),
				t.Format(time.RFC3339),
			))
		}

		writeEvent := func(event string, data interface{}) {
			log.Println("write event")
			buf := &bytes.Buffer{}
			if err := json.NewEncoder(buf).Encode(data); err != nil {
				log.Println(err)
				return
			}

			io.WriteString(w, fmt.Sprintf(
				`id: %d
event: %s
data: %s 

`,
				time.Now().Unix(),
				event,
				buf.String(),
			))
			if _, err := w.Write(buf.Bytes()); err != nil {
				log.Println(err)
				return
			}
		}

		connClosed := w.(http.CloseNotifier).CloseNotify()
		flusher := w.(http.Flusher)
		writeMsg(time.Now())
		// Output first timestamp
		flusher.Flush()

		interval := time.NewTicker(5 * time.Second)
	loop:
		for {
			select {
			case <-connClosed:
				interval.Stop()
				break loop
			case t := <-interval.C:
				writeMsg(t)
				flusher.Flush()
			case msg := <-msgChannel:
				type chatMessage struct {
					Msg string `json:"msg"`
				}
				writeEvent("chat_message", chatMessage{
					Msg: msg,
				})
				flusher.Flush()
			}
		}
	}))

	if err := http.ListenAndServe(":3000", mux); err != nil {
		log.Fatal(err)
	}
}

func postMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if msg := r.Form.Get("msg"); msg != "" {
		log.Println("passing msg into chan " + msg)
		msgChannel <- msg
	}

	w.WriteHeader(http.StatusNoContent)
}
