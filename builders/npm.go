package builders

import (
	"io"
	"os"
	"os/exec"
)

func npm(sourceDir string, output io.Writer) error {
	// Use NPM by default. But if yarn.lock exists, use Yarn.
	cmd := exec.Command("npm", "install")
	if exists(sourceDir + "/yarn.lock") {
		cmd = exec.Command("yarn")
	}
	cmd.Dir = sourceDir
	cmd.Stderr = output
	cmd.Stdout = output
	return cmd.Run()
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
