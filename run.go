package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func run(cwd string, prog string, args ...string) error {
	log.Println(cwd, prog, args)
	c := exec.Command(prog, args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	c.Dir = cwd
	c.Env = os.Environ()
	return c.Run()
}

func runOut(cwd string, prog string, args ...string) (string, error) {
	c := exec.Command(prog, args...)
	c.Stderr = os.Stderr
	c.Dir = cwd
	out, err := c.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func npm(sourceDir string) error {
	if exists(sourceDir + "/yarn.lock") {
		return run(sourceDir, "yarn")
	}
	return run(sourceDir, "npm install")
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
