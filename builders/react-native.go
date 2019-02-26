package builders

import "io"

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
func (b *ReactNativeBuilder) Build(output io.Writer) ([]string, error) {
	var err error
	err = npm(b.projectDir, output)
	if err != nil {
		return nil, err
	}
	paths, err := b.android.Build(output)
	return paths, err
}

// Name returns the builder's name.
func (b *ReactNativeBuilder) Name() string {
	return "React Native"
}

// Dirname returns the builder's project path.
func (b *ReactNativeBuilder) Dirname() string {
	return b.projectDir
}
