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

func buildLogger(projectName, branch, sourceID string) (io.WriteCloser, error) {
	logPath := "projects/" + projectName + "/builds/" + safeString(branch) + "/" + safeString(sourceID) + ".log"
	err := os.MkdirAll(path.Dir(logPath), 0777)
	if err != nil {
		return nil, err
	}
	return os.Create(logPath)
}

// publish publishes the given array of builds on the webserver.
func publish(projectDir, branch, sourceID string, files []string) error {
	pubDir := fmt.Sprintf("%s/builds/%s/%s", projectDir, safeString(branch), safeString(sourceID))
	return copyFiles(files, pubDir)
}

func branchesList(projectName string) ([]string, error) {
	dirs, err := ls("projects/" + projectName + "/builds")
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

func buildsList(projectName, branch string) ([]*build, error) {
	buildDirs, err := ls("projects/" + projectName + "/builds/" + safeString(branch))
	if err != nil {
		return nil, err
	}

	result := make([]*build, 0)
	for _, path := range buildDirs {
		files, err := ls(path)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			result = append(result, &build{
				path: file,
			})
		}
	}
	return result, nil
}

func isSafeChar(ch rune) bool {
	return (ch >= '0' && ch <= '9') ||
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

func ls(dir string) ([]string, error) {
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
