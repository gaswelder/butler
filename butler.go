package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	go trackUpdates()
	go serveBuilds()
	select {}
}

// trackUpdates continuously updates the projects directory
// and makes new builds.
func trackUpdates() {
	for {
		projects, err := listProjects()
		if err != nil {
			// If something went wrong while trying to get the list of
			// projects, wait a little before trying again.
			log.Printf("failed to get projects list: %v", err)
			time.Sleep(60 * time.Second)
			continue
		}

		for _, project := range projects {
			log.Printf("Project: %s", project.name)
			err := project.update()
			if err != nil {
				log.Fatal(err)
			}
		}

		// Sleep a while and repeat the whole thing again.
		time.Sleep(10 * time.Second)
	}
}

// serveBuilds spawns an HTTP server that serves all builds for all projects.
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
