package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

func isValidName(name string) bool {
	for _, ch := range name {
		if ch == '.' {
			return false
		}
		ok := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
		if !ok {
			return false
		}
	}
	return true
}

// serveBuilds spawns an HTTP server that serves all builds for all projects.
func serveBuilds() {
	// /projectName/master/
	//     1.1.0-dev
	//     1.1.0-prod
	//     1.1.0-staging
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.RequestURI, "/", 3)
		projectName := parts[1]
		branch := parts[2]
		if !isValidName(projectName) || !isValidName(branch) {
			w.WriteHeader(400)
			fmt.Fprintf(w, "invalid path")
			return
		}

		builds, err := buildsList(projectName, branch)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "%s", err.Error())
			return
		}

		w.Header().Add("Content-Type", "text/html;charset=utf-8")
		fmt.Fprintf(w, "<h1>%s, %s</h1>", projectName, branch)
		for _, build := range builds {
			fmt.Fprintf(w, "<li>%s", build.url())
		}
	})
	http.ListenAndServe(":8080", nil)
}

// publish publishes the given array of builds on the webserver.
func publish(projectDir, branch, sourceID string, files []string) error {
	pubDir := fmt.Sprintf("%s/builds/%s/%s", projectDir, branch, sourceID)
	return copyFiles(files, pubDir)
}

func buildsList(projectName, branch string) ([]*build, error) {
	buildDirs, err := ls("projects/" + projectName + "/builds/" + branch)
	if err != nil {
		return nil, err
	}

	result := make([]*build, 0)
	for _, path := range buildDirs {
		files, err := ls(path)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			result = append(result, &build{
				path: file,
			})
		}
	}
	return result, nil
}

func copyFiles(paths []string, to string) error {
	err := os.MkdirAll(to, 0777)
	if err != nil {
		return err
	}
	for _, f := range paths {
		err := exec.Command("cp", f, to+"/"+path.Base(f)).Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func ls(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ls, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(ls))
	for i, l := range ls {
		result[i] = dir + "/" + l.Name()
	}
	return result, nil
}
