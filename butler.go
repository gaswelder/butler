package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	go trackReps()
	go serveBuilds()
	select {}
}

func trackReps() {
	for {
		projects, err := listProjects()
		if err != nil {
			log.Printf("failed to get projects list: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, project := range projects {
			log.Printf("Project: %v", project)
			err := project.update()
			if err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func serveBuilds() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hey there")
	})
	http.HandleFunc("/builds/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.RequestURI, "/", 3)
		projectName := parts[2]
		if projectName == "" {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Missing project name in the URL")
			return
		}
		project := getProject(projectName)
		if project == nil {
			w.WriteHeader(404)
			fmt.Fprintf(w, "Unknown project: %s", projectName)
		}

		builds, err := project.builds()
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, err.Error())
			return
		}
		for _, build := range builds {
			fmt.Fprintf(w, build.url())
		}
	})
	http.ListenAndServe(":8080", nil)
}
