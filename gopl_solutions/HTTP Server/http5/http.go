package main

import(
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type fileHandler struct{
	root string
	hand http.Handler
}

// ServerHTTP implement the http.Handler interface
func (h *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	log.Println(r.Method, r.URL.Path)
	name := filepath.Join(h.root, r.URL.Path)
	switch r.Method{
	case "GET","HEAD":
		h.serverGet(w, r, name)
	case "PUT","POST":
		h.servePut(w, r, name)
	default :
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h *fileHandler) serverGet(w http.ResponseWriter, r *http.Request, name string){
	d, err := os.Stat(name)
	if err != nil || d.IsDir(){
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	f, err := os.Open(name)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	http.ServeContent(w, r, name, d.ModTime(), f)
}

func (h *fileHandler) servePut(w http.ResponseWriter, r *http.Request, name string){
	d, err := os.Stat(name)
	if err == nil && d.IsDir(){
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _,err := io.Copy(f, r.Body); err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func FileServer(root string) http.Handler{
	return &fileHandler{root : root, hand:http.FileServer(http.Dir(root))}
}

func main(){
	flPort := flag.String("port", ":8080", "port")
	flag.Parse()
	if err := http.ListenAndServe(*flPort, FileServer(flag.Arg(0))); err != nil{
		log.Fatal("ListenAndServe: ", err)
	}
}