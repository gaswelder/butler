package builders

import "os"

// Builder represents a builder object for a particular kind of project.
type Builder interface {
	// Build performs a build and returns a list of output file paths.
	Build() ([]string, error)

	// Dirname returns the builder's project path.
	Dirname() string

	// Name returns the builder's name.
	Name() string
}

// Find determines and returns the appropriate builder for the given project root.
func Find(sourceDir string) (Builder, error) {
	var ok bool
	var err error

	ok, err = hasFiles(sourceDir, "package.json", "android", "ios")
	if err != nil {
		return nil, err
	}
	if ok {
		return ReactNative(sourceDir), nil
	}

	ok, err = hasFiles(sourceDir, "gradlew", ".project")
	if err != nil {
		return nil, err
	}
	if ok {
		return Android(sourceDir), nil
	}
	return nil, nil
}

func hasFiles(sourceDir string, files ...string) (bool, error) {
	for _, name := range files {
		path := sourceDir + "/" + name
		_, err := os.Stat(path)
		if err == nil {
			continue
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
