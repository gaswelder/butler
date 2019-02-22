package builders

// ReactNativeBuilder is a buider for React Native projects.
type ReactNativeBuilder struct {
	projectDir string
	android    Builder
}

// ReactNative returns a builder for a React Native project.
func ReactNative(projectDir string) Builder {
	return &ReactNativeBuilder{
		projectDir: projectDir,
		android:    Android(projectDir + "/android"),
	}
}

// Build builds the project.
func (b *ReactNativeBuilder) Build() ([]string, error) {
	var err error
	err = npm(b.projectDir)
	if err != nil {
		return nil, err
	}
	paths, err := b.android.Build()
	return paths, err
}
