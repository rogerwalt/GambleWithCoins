// received map is empty. doesn't work.
// use e.g. with $.ajax({type: "POST", url: "http://localhost:8080/jsonhere", data: {foo: "bar"}})

package main

import (
    "fmt"
    "net/http"
    "os"
    "encoding/json"
    "io/ioutil"
)

func main() {
	// receive json
	http.HandleFunc("/jsonhere/", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		contents, err := ioutil.ReadAll(req.Body)
		checkError(err)
		fmt.Println(string(contents))

		var m map[string]interface{}
		json.Unmarshal(contents, &m)

		fmt.Printf("json received: %v", m)
	})

	// serve static files
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, req.URL.Path[1:])
	})

	// start webserver
	err := http.ListenAndServe(":8080", nil)

	// check for errors in webserver
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}