package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("listening on port 80")
	http.Handle("/", http.FileServer(http.Dir("../db/datas")))
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("server down")
}
