package server

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

// Publish publishes the given array of builds on the webserver.
func Publish(projectDir, branch, sourceID string, files []string) error {
	pubDir := fmt.Sprintf("%s/builds/%s/%s", projectDir, branch, sourceID)
	return copyFiles(files, pubDir)
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
