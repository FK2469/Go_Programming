package main

import "net/http"

func HandlerIndex(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("Hello World!"))
}

func main(){
	http.HandleFunc("/", HandlerIndex)
	http.ListenAndServe(":8080", nil)
}