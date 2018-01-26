//静态文件服务器
package main

import(
	"log"
	"net/http"
)

func main(){
	http.Handle("/", http.FileServer(http.Dir("/tmp")))
	if err := http.ListenAndServe(":80", nil); err != nil{
		log.Fatal("ListenAndServe: ", err)
	}
}