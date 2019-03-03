package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gaswelder/butler/storage"
)

// serveBuilds spawns an HTTP server that serves all builds for all projects.
func serveBuilds() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.RequestURI, "/")
		if parts[0] == "" {
			parts = parts[1:]
		}
		if len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
		for _, p := range parts {
			if !isValidName(p) {
				statusPage(w, 400, "Invalid URL")
				return
			}
		}

		n := len(parts)
		if n == 0 {
			rootPage(w)
			return
		}
		if n == 1 {
			projectIndex(w, parts[0])
			return
		}
		if n == 2 {
			branchIndex(w, parts[0], parts[1])
			return
		}
		if n == 3 {
			versionIndex(w, parts[0], parts[1], parts[2])
			return
		}
		if n == 4 {
			serveBuild(w, parts[0], parts[1], parts[2], parts[3])
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
	branches, err := storage.Branches(projectName)
	if err != nil {
		statusPage(w, 500, err.Error())
		return
	}

	w.Header().Add("Content-Type", "text/html;charset=utf-8")
	fmt.Fprintf(w, "<h1>%s</h1>", projectName)
	fmt.Fprint(w, breadcrumbs(projectName))
	fmt.Fprint(w, "<ol>")
	for _, branch := range branches {
		fmt.Fprintf(w, "<li><a href=\"/%s/%s/\">%s</a></li>", projectName, branch, branch)
	}
	fmt.Fprint(w, "</ol>")
}

func branchIndex(w http.ResponseWriter, projectName, branch string) {
	versions, err := storage.Versions(projectName, branch)
	if err != nil {
		statusPage(w, 500, "Failed to get versions list: "+err.Error())
		return
	}
	w.Header().Add("Content-Type", "text/html;charset=utf-8")
	fmt.Fprintf(w, "<h1>%s</h1>", projectName)
	fmt.Fprint(w, breadcrumbs(projectName, branch))
	fmt.Fprint(w, "<ol>")
	for _, v := range versions {
		fmt.Fprintf(w, "<li><a href=\"/%s/%s/%s\">%s</a></li>", projectName, branch, v, v)
	}
	fmt.Fprint(w, "</ol>")
}

func versionIndex(w http.ResponseWriter, projectName, branch, version string) {
	builds, err := storage.Builds(projectName, branch, version)
	if err != nil {
		statusPage(w, 500, "Failed to get builds list: "+err.Error())
		return
	}
	w.Header().Add("Content-Type", "text/html;charset=utf-8")
	fmt.Fprintf(w, "<h1>%s</h1>", projectName)
	fmt.Fprint(w, breadcrumbs(projectName, branch, version))
	fmt.Fprint(w, "<ol>")
	for _, b := range builds {
		fmt.Fprintf(w, "<li><a href=\"/%s/%s/%s/%s\">%s</a></li>", projectName, branch, version, b, b)
	}
	fmt.Fprint(w, "</ol>")
}

func serveBuild(w http.ResponseWriter, project, branch, version, file string) {
	f, err := storage.Build(project, branch, version, file)
	if os.IsNotExist(err) {
		statusPage(w, 404, "Not found")
		return
	}
	if err != nil {
		statusPage(w, 500, err.Error())
		return
	}
	defer f.Close()
	if strings.HasSuffix(file, ".log") {
		w.Header().Add("Content-Type", "text/plain;charset=utf-8")
	}
	io.Copy(w, f)
}

func statusPage(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	fmt.Fprint(w, message)
}

func isValidName(name string) bool {
	for i, ch := range name {
		if i == 0 && ch == '.' {
			return false
		}
		if ch == '/' {
			return false
		}
	}
	return true
}

func breadcrumbs(parts ...string) string {
	b := strings.Builder{}
	b.WriteString("<nav>")
	for i, v := range parts[:len(parts)-1] {
		path := strings.Join(parts[:i+1], "/")
		b.WriteString("<a href=\"/" + path + "\">" + v + "</a> / ")
	}
	b.WriteString(parts[len(parts)-1])
	b.WriteString("</nav>")
	return b.String()
}
