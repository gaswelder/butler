package main

import "strings"

type git struct {
	sourceDir string
}

func (g git) pull() error {
	return run(g.sourceDir, "git", "pull")
}

// branches returns a list of remote branches in this repository.
func (g git) branches() ([]string, error) {
	out, err := runOut(g.sourceDir, "git", "branch", "-r")
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.Index(line, " -> ") > 0 {
			continue
		}
		parts := strings.SplitN(line, "/", 2)
		names = append(names, parts[1])
	}
	return names, nil
}

// checkout checks out the given branch.
func (g git) checkout(branch string) error {
	return run(g.sourceDir, "git", "checkout", branch)
}

// describe returns the output of "git describe" on the current branch.
func (g git) describe() (string, error) {
	return runOut(g.sourceDir, "git", "describe")
}
