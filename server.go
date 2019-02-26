package main

import (
	"fmt"
	"net/http"
	"strings"
)

// serveBuilds spawns an HTTP server that serves all builds for all projects.
func serveBuilds() {
	// /projectName/master/
	//     1.1.0-dev
	//     1.1.0-prod
	//     1.1.0-staging
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.RequestURI, "/")
		if parts[0] == "" {
			parts = parts[1:]
		}
		if len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
		n := len(parts)
		if n == 0 {
			rootPage(w)
			return
		}
		projectName := parts[0]
		if !isValidName(projectName) {
			statusPage(w, 400, "Invalid project name")
			return
		}
		if n == 1 {
			projectIndex(w, projectName)
			return
		}
		branch := parts[1]
		if !isValidName(branch) {
			statusPage(w, 400, "Invalid branch name: "+branch)
			return
		}
		if n == 2 {
			branchIndex(w, projectName, branch)
			return
		}
		statusPage(w, 404, "Not found")
	})
	http.ListenAndServe(":8080", nil)
}

func rootPage(w http.ResponseWriter) {
	statusPage(w, 200, "Nothing here")
}

func projectIndex(w http.ResponseWriter, projectName string) {
	branches, err := branchesList(projectName)
	if err != nil {
		statusPage(w, 500, err.Error())
		return
	}

	w.Header().Add("Content-Type", "text/html;charset=utf-8")
	fmt.Fprintf(w, "<h1>%s branches</h1>", projectName)
	for _, branch := range branches {
		fmt.Fprintf(w, "<li><a href=\"/%s/%s/\">%s</a></li>", projectName, branch, branch)
	}
}

func branchIndex(w http.ResponseWriter, projectName, branch string) {
	builds, err := buildsList(projectName, branch)
	if err != nil {
		statusPage(w, 500, err.Error())
		return
	}
	w.Header().Add("Content-Type", "text/html;charset=utf-8")
	fmt.Fprintf(w, "<h1>%s, %s</h1>", projectName, branch)
	for _, build := range builds {
		fmt.Fprintf(w, "<li>%s", build.url())
	}
}

func statusPage(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	fmt.Fprint(w, message)
}

func isValidName(name string) bool {
	for _, ch := range name {
		if ch == '.' || ch == '/' {
			return false
		}
	}
	return true
}
