package main

import (
	"fmt"
	"os"
)

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
