package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func createCluster(name string) {
	cmd := exec.Command("task", fmt.Sprintf("build:%s", name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func destroyCluster(name string) {
	cmd := exec.Command("task", fmt.Sprintf("destroy:%s", name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	//	createCluster("poc-cluster")
	//	createCluster("dev-cluster")
	//	createCluster("prod-cluster")
	//	destroyCluster("poc-cluster")
	//	destroyCluster("dev-cluster")
	//	destroyCluster("prod-cluster")

	http.HandleFunc("/build", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		fmt.Fprintf(w, "building %s", name)
		if len(name) > 0 {
			go createCluster(name)
			return
		}
		fmt.Fprintf(w, "No name")
	})

	http.HandleFunc("/destroy", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		fmt.Fprintf(w, "destroying %s", name)
		if len(name) > 0 {
			go destroyCluster(name)
			return
		}
		fmt.Fprintf(w, "No name")
	})

	s := http.Server{
		Addr:           ":8090",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()
}
