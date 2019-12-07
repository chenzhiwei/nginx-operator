package main

import (
	"fmt"
	"io"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, this message is from Fake Nginx!\n")
}

func main() {
	fmt.Println("Fake Nginx Server starts at port 8080")
	http.HandleFunc("/", helloHandler)
	http.ListenAndServe(":8080", nil)
}
