package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/mux"
)

const defaultTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>
	<body>
		<h1>Live</h1>
		<p>{{.Len}} viewers when page loaded</p>
		<img src="{{.Stream}}"/>
	</body>
</html>`

func init() {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			debug.FreeOSMemory()
		}
	}()
}

func main() {
	fmt.Println("http://localhost:8080")
	Broadcast()
	router := mux.NewRouter()
	router.HandleFunc("/", getTemplate).Methods("GET")
	router.HandleFunc("/stream", GetMJPEG).Methods("GET")
	http.ListenAndServe(":8080", router)
}

// handles request GET getTemplate
func getTemplate(w http.ResponseWriter, r *http.Request) {
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	t, err := template.New("webpage").Parse(defaultTemplate)
	check(err)

	data := struct {
		Title  string
		Len    int
		Stream string
	}{
		Title:  "MJPG Server",
		Stream: "/stream",
	}
	data.Len = Len()
	err = t.Execute(w, data)
	check(err)
}

// GetMJPEG handles request GET GetMJPEG
func GetMJPEG(w http.ResponseWriter, r *http.Request) {
	WriteStreamOutput(w)
}
