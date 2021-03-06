package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"
)

// ReleasesDirectory is a special directory where builds with clean tags are stored.
const ReleasesDirectory = "releases"

// Project represents a project to be built.
type Project struct {
	Name string
	Env  []string
}

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

// Has returns true if there are saved results for given project, branch and version.
func Has(project, branch, version string) bool {
	path := buildsPath(project, branch, version)
	_, err := os.Stat(path)
	return err == nil
}

func parseDotEnv(path string) ([]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	list := make([]string, 0)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		list = append(list, line)
	}
	return list, nil
}

// Projects returns the current list of projects.
func Projects() ([]Project, error) {
	dirs, err := lsd("projects")
	if err != nil {
		return nil, err
	}

	projects := make([]Project, len(dirs))
	for i, dir := range dirs {
		env, err := parseDotEnv(dir + "/.env")
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		projects[i] = Project{
			Name: path.Base(dir),
			Env:  env,
		}
	}
	return projects, nil
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
	r := baseNames(l)
	sort.Sort(sort.Reverse(sort.StringSlice(r)))
	return r, nil
}

// Builds returns a list of builds.
func Builds(project, branch, version string) ([]string, error) {
	files, err := lsf(buildsPath(project, branch, version))
	if err != nil {
		return nil, err
	}
	r := baseNames(files)
	sort.Strings(r)
	return r, nil
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
