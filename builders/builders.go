package builders

// Builder represents a builder object for a particular kind of project.
type Builder interface {
	// Build performs a build and returns a list of output file paths.
	Build() ([]string, error)
}

// Find determines and returns the appropriate builder for the given project root.
func Find(sourceDir string) (Builder, error) {
	// Just return a React Native builder for now.
	return ReactNative(sourceDir), nil
}
