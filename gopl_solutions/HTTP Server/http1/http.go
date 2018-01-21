// Handle "POST"
package main

import(
	"io/ioutil"
	"net/http"
)

func HandlerPost(w http.ResponseWriter, r *http.Request){
	if r.Method == "POST"{
		content, err := ioutil.ReadAll(r.Body)
		if err != nil{
			w.WriteHeader(400)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(200)
		w.Write(content)
		return
	}
}

func main(){
	http.HandleFunc("/post", HandlerPost)
	http.ListenAndServe("127.0.0.1:8080", nil)
}