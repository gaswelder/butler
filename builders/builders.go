package builders

// Builder represents a builder object for a particular kind of project.
type Builder interface {
	// Build performs a build and returns a list of output file paths.
	Build() ([]string, error)
}
