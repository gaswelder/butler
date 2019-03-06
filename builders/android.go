package builders

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

// AndroidBuilder is a builder for Gradle-based Android projects.
type AndroidBuilder struct {
	projectDir string
}

// Android returns an Android builder.
func Android(projectDir string) Builder {
	return &AndroidBuilder{
		projectDir: projectDir,
	}
}

// Build builds the project.
func (a *AndroidBuilder) Build(output io.Writer, envVars []string) ([]string, error) {
	projectDir := a.projectDir

	cmd := exec.Command("./gradlew", "build")
	cmd.Dir = projectDir
	cmd.Stderr = output
	cmd.Stdout = output
	cmd.Env = append(envVars)

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return findFiles(projectDir+"/app/build/outputs/apk", ".apk")
}

// Recursively scands the given directory and returns all paths
// to files having the given suffix.
func findFiles(dir, suffix string) ([]string, error) {
	s, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	ls, err := s.Readdir(-1)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0)
	for _, l := range ls {
		if l.IsDir() {
			more, err := findFiles(dir+"/"+l.Name(), suffix)
			if err != nil {
				return nil, err
			}
			paths = append(paths, more...)
			continue
		}
		if strings.HasSuffix(l.Name(), suffix) {
			paths = append(paths, dir+"/"+l.Name())
			continue
		}
	}

	return paths, nil
}

// Name returns the builder's name.
func (a *AndroidBuilder) Name() string {
	return "Android"
}

// Dirname returns the builder's project path.
func (a *AndroidBuilder) Dirname() string {
	return a.projectDir
}
