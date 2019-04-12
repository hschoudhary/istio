package main

import (
	"fmt"
    "log"
	"net/http"
	"sync"
)

type requestCount struct {
	requestCount int64
	mu *sync.Mutex
}
func main() {

	req := requestCount{
		0, &sync.Mutex{},
	}
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    	fmt.Sprintf("Received request, requester=%s", r.Header.Get("User-Agent"))
    	http.Error(w, "***Unsupported endpoint***", 503)
    })

    http.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
    	fmt.Sprintf("Received request, requester=%s", r.Header.Get("User-Agent"))
    	req.mu.Lock()
    	req.requestCount++
    	fmt.Fprintf(w, "!! Awesome Demo!! \n Demo request number=%d", req.requestCount)
    	req.mu.Unlock()
    })

    log.Fatal(http.ListenAndServe(":8090", nil))

}
