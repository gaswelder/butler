package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// SourcePath returns path to a project's source directory.
func SourcePath(project string) string {
	return "projects/" + project + "/src"
}

func versionsPath(project, branch string) string {
	return "projects/" + project + "/builds/" + safeString(branch)
}

func buildsPath(project, branch, version string) string {
	return versionsPath(project, branch) + "/" + safeString(version)
}

// ProjectsList returns the current list of projects.
func ProjectsList() ([]string, error) {
	dirs, err := lsd("projects")
	if err != nil {
		return nil, err
	}
	return baseNames(dirs), nil
}

// BuildLogger creates and returns a log writer for a build process.
func BuildLogger(project, branch, version string) (io.WriteCloser, error) {
	logPath := versionsPath(project, branch) + "/" + safeString(version) + "/build.log"
	err := os.MkdirAll(path.Dir(logPath), 0777)
	if err != nil {
		return nil, err
	}
	return os.Create(logPath)
}

// Build returns a reader for a build file.
func Build(project, branch, version, file string) (io.ReadCloser, error) {
	return os.Open(buildsPath(project, branch, version) + "/" + safeString(file))
}

// SaveBuilds stores build outputs for the given project, branch and version.
func SaveBuilds(project, branch, version string, files []string) error {
	err := copyFiles(files, buildsPath(project, branch, version))
	if err != nil {
		return fmt.Errorf("failed to copy files: %v", err)
	}
	// Update the latest mark.
	err = SetLatestBuildID(project, branch, version)
	if err != nil {
		return fmt.Errorf("failed to update version ID: %v", err)
	}
	return nil
}

// Branches returns a list of project's branches.
func Branches(project string) ([]string, error) {
	dirs, err := lsd("projects/" + project + "/builds")
	if err != nil {
		return nil, err
	}

	branches := make([]string, len(dirs))
	for i, dir := range dirs {
		branches[i] = path.Base(dir)
	}
	return branches, nil
}

// LatestVersion returns latest saved version for the given project and branch.
func LatestVersion(project, branch string) (string, error) {
	path := fmt.Sprintf("projects/%s/latest-%s", project, safeString(branch))
	b, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return "(no build)", nil
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// SetLatestBuildID saves the version ID of the latest built source.
func SetLatestBuildID(projectName, branch, id string) error {
	path := fmt.Sprintf("projects/%s/latest-%s", projectName, safeString(branch))
	return ioutil.WriteFile(path, []byte(id), 0666)
}

func baseNames(paths []string) []string {
	names := make([]string, len(paths))
	for i, p := range paths {
		names[i] = path.Base(p)
	}
	return names
}

// Versions returns a list of build versions.
func Versions(project, branch string) ([]string, error) {
	l, err := lsd(versionsPath(project, branch))
	if err != nil {
		return nil, err
	}
	return baseNames(l), nil
}

// Builds returns a list of builds.
func Builds(project, branch, version string) ([]string, error) {
	files, err := lsf(buildsPath(project, branch, version))
	if err != nil {
		return nil, err
	}
	return baseNames(files), nil
}

func isSafeChar(ch rune) bool {
	return ch == '.' ||
		(ch >= '0' && ch <= '9') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= 'a' && ch <= 'z')
}

func safeString(name string) string {
	b := strings.Builder{}
	for _, ch := range name {
		if !isSafeChar(ch) {
			b.WriteRune('-')
			continue
		}
		b.WriteRune(ch)
	}
	return b.String()
}

func lsd(dir string) ([]string, error) {
	return ls(dir, func(f os.FileInfo) bool {
		return f.IsDir()
	})
}

func lsf(dir string) ([]string, error) {
	return ls(dir, func(f os.FileInfo) bool {
		return !f.IsDir()
	})
}

func ls(dir string, filter func(f os.FileInfo) bool) ([]string, error) {
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
		if !filter(l) {
			continue
		}
		result[i] = dir + "/" + l.Name()
	}
	return result, nil
}

// Stash copies the given files to a temporary place adding envName to their names.
func Stash(files []string, envName string) ([]string, error) {
	r := make([]string, len(files))
	for i, f := range files {
		nonce := time.Now().UnixNano()
		stashPath := fmt.Sprintf("./tmp/%d/%s-%s", nonce, envName, path.Base(f))
		err := os.MkdirAll(path.Dir(stashPath), 0777)
		if err != nil {
			return nil, err
		}
		err = copyFile(f, stashPath)
		if err != nil {
			return nil, err
		}
		r[i] = stashPath
	}
	return r, nil
}

func copyFile(from, to string) error {
	return exec.Command("cp", from, to).Run()
}

func copyFiles(paths []string, to string) error {
	err := os.MkdirAll(to, 0777)
	if err != nil {
		return err
	}
	for _, f := range paths {
		err := copyFile(f, to+"/"+path.Base(f))
		if err != nil {
			return err
		}
	}
	return nil
}
