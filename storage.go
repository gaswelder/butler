package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

func versionsPath(project, branch string) string {
	return "projects/" + project + "/builds/" + safeString(branch)
}

func buildsPath(project, branch, version string) string {
	return versionsPath(project, branch) + "/" + safeString(version)
}

func buildLogger(projectName, branch, sourceID string) (io.WriteCloser, error) {
	logPath := versionsPath(projectName, branch) + "/" + safeString(sourceID) + "/build.log"
	err := os.MkdirAll(path.Dir(logPath), 0777)
	if err != nil {
		return nil, err
	}
	return os.Create(logPath)
}

func buildFile(project, branch, version, file string) (io.ReadCloser, error) {
	return os.Open(buildsPath(project, branch, version) + "/" + safeString(file))
}

// publish publishes the given array of builds on the webserver.
func publish(projectDir, branch, sourceID string, files []string) error {
	pubDir := fmt.Sprintf("%s/builds/%s/%s", projectDir, safeString(branch), safeString(sourceID))
	return copyFiles(files, pubDir)
}

func branchesList(projectName string) ([]string, error) {
	dirs, err := lsd("projects/" + projectName + "/builds")
	if err != nil {
		return nil, err
	}

	branches := make([]string, len(dirs))
	for i, dir := range dirs {
		branches[i] = path.Base(dir)
	}
	return branches, nil
}

func latestBuildID(projectName, branch string) (string, error) {
	path := fmt.Sprintf("projects/%s/latest-%s", projectName, safeString(branch))
	b, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return "(no build)", nil
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func setLatestBuildID(projectName, branch, id string) error {
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

func versionsList(project, branch string) ([]string, error) {
	l, err := lsd(versionsPath(project, branch))
	if err != nil {
		return nil, err
	}
	return baseNames(l), nil
}

func buildsList(project, branch, version string) ([]string, error) {
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
