// 静态文件服务器，自定义前缀
package main

import(
	"log"
	"net/http"
)

func main(){
	http.Handle("/fs/", http.StripPrefix("/fs/", http.FileServer(http.Dir("/tmp"))))
	if err := http.ListenAndServe(":80", nil); err != nil{
		log.Fatal("ListenAndServe: ", err)
	}
}