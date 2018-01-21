// 创建私有路由(默认情况下,使用 http.DefaultServeMux)
package main

import(
	"log"
	"net/http"
	"time"
)

type ServeMux struct{
	http.ServeMux
}

func NewServeMux() *ServeMux{
	return &ServeMux{ServeMux: *http.NewServeMux()}
}

func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request){
	date := time.Now()
	defer func(){
		r.Body.Close()
		log.Printf("%.2f %s %s \n", float64(time.Since(date).Nanoseconds()) / 1000000, r.Method, r.URL)
	}()
	mux.ServeMux.ServeHTTP(w,r)
}

func HandlerIndex(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("Hello World!"))
}

func main(){
	mux := NewServeMux()
	mux =.HandleFunc("/", HandlerIndex)
	sever := http.Server{
		Addr : ":80",
		Handler : mux,
	}
	if err := server.ListenAndServe(); err != nil{
		log.Fatalln(err)
	}
}