package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

type build struct {
	path string
}

func (b *build) url() string {
	return b.path
}

type project struct {
	name string
	dir  string
}

// listProjects returns the current list of projects.
func listProjects() ([]*project, error) {
	f, err := os.Open("projects")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ls, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	result := make([]*project, len(ls))
	for i, l := range ls {
		result[i] = &project{
			name: l.Name(),
			dir:  fmt.Sprintf("projects/%s", l.Name()),
		}
	}
	return result, nil
}

func getProject(name string) *project {
	dir := "projects/" + name
	if !exists(dir) {
		return nil
	}
	return &project{
		name: name,
		dir:  dir,
	}
}

// builds returns a list of the project's builds.
func (p *project) builds() ([]*build, error) {
	f, err := os.Open(p.dir + "/builds")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ls, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	result := make([]*build, len(ls))
	for i, l := range ls {
		result[i] = &build{
			path: p.dir + "/builds/" + l.Name(),
		}
	}
	return result, nil
}

func (p *project) latestBuildID(branch string) (string, error) {
	path := fmt.Sprintf("%s/latest-%s", p.dir, branch)
	b, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return "(no build)", nil
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (p *project) setLatestBuildID(branch, id string) error {
	path := fmt.Sprintf("%s/latest-%s", p.dir, branch)
	return ioutil.WriteFile(path, []byte(id), 0666)
}
