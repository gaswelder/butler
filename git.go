package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type git struct {
	sourceDir string
}

func run(cwd string, prog string, args ...string) error {
	c := exec.Command(prog, args...)
	c.Dir = cwd
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func runOut(cwd string, prog string, args ...string) ([]string, error) {
	cmd := exec.Command(prog, args...)
	cmd.Dir = cwd
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

func (g git) pull() error {
	return run(g.sourceDir, "git", "pull")
}

func (g git) fetch() error {
	err := run(g.sourceDir, "git", "fetch", "-p")
	if err != nil {
		return err
	}
	err = g.pruneTags()
	return err
}

// Takes an output of git show-ref or git ls-remote as an array of lines
// and returns an array of extracted tag names.
func parseTags(lines []string) ([]string, error) {
	n := len("refs/tags/")
	tags := make([]string, len(lines))
	for i, line := range lines {
		pos := strings.Index(line, "refs/tags/")
		if pos < 0 {
			return nil, fmt.Errorf("failed to parse tag line: %s", line)
		}
		tags[i] = line[pos+n:]
	}
	return tags, nil
}

// Returns an array of strings that are present in a, but not b.
func subtractArray(a, b []string) []string {
	has := func(l []string, x string) bool {
		for _, v := range l {
			if v == x {
				return true
			}
		}
		return false
	}
	diff := make([]string, 0)
	for _, v := range a {
		if !has(b, v) {
			diff = append(diff, v)
		}
	}
	return diff
}

func (g git) pruneTags() error {
	// Get remote tags and local tags.
	remote, err := runOut(g.sourceDir, "git", "ls-remote", "-t")
	if err != nil {
		return err
	}
	remote, err = parseTags(remote)
	if err != nil {
		return err
	}
	local, err := runOut(g.sourceDir, "git", "show-ref", "--tags")
	if err != nil {
		return err
	}
	local, err = parseTags(local)
	if err != nil {
		return err
	}

	// Figure out what tags were deleted on the server and delete them locally.
	deleted := subtractArray(local, remote)
	for _, tag := range deleted {
		fmt.Fprintf(os.Stdout, "deleting tag %s\n", tag)
		err := run(g.sourceDir, "git", "tag", "-d", tag)
		if err != nil {
			return fmt.Errorf("failed to delete tag %s: %v", tag, err)
		}
	}

	return nil
}

func (g git) discard() error {
	err := run(g.sourceDir, "git", "clean", "-fd")
	if err != nil {
		return err
	}
	return run(g.sourceDir, "git", "checkout", ".")
}

type branch struct {
	name string
	desc string
}

func branchIsBuildable(branch string) bool {
	return strings.HasPrefix(branch, "dev") || branch == "master" || branch == "butler"
}

func (g git) tags() ([]string, error) {
	lines, err := runOut(g.sourceDir, "git", "tag", "-l", "--sort", "v:refname")
	if err != nil {
		return nil, err
	}

	versions := make([]string, 0)
	for _, line := range lines {
		isTag, err := regexp.Match(`^\d+.\d+.\d+(-\d+)?$`, []byte(line))
		if err != nil {
			return nil, fmt.Errorf("regexp error: %v", err)
		}
		if !isTag {
			continue
		}
		versions = append(versions, line)
	}
	return versions, nil
}

// branches returns a list of remote branches in this repository.
func (g git) branches() ([]branch, error) {
	lines, err := runOut(g.sourceDir, "git", "branch", "-r")
	if err != nil {
		return nil, err
	}

	branches := make([]branch, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Index(line, " -> ") > 0 {
			continue
		}

		// Split "origin/develop" to ["origin", "develop"].
		parts := strings.SplitN(line, "/", 2)

		// Skip branches that we are not going to build.
		if !branchIsBuildable(parts[1]) {
			continue
		}
		desc, err := g.describe(line)
		if err != nil {
			return nil, err
		}

		branches = append(branches, branch{name: parts[1], desc: desc})
	}
	return branches, nil
}

// checkout checks out the given branch.
func (g git) checkout(branch string) error {
	return run(g.sourceDir, "git", "checkout", branch)
}

// describe returns the output of "git describe" on the given ref
func (g git) describe(ref string) (string, error) {
	lines, err := runOut(g.sourceDir, "git", "describe", "--tags", ref)
	if err != nil {
		return "", err
	}
	if len(lines) != 1 {
		return "", fmt.Errorf("describe: wrong output lines count (%v)", lines)
	}
	return lines[0], nil
}
