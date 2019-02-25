package builders

import "os"

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
func (a *AndroidBuilder) Build() ([]string, error) {
	projectDir := a.projectDir

	err := run(projectDir, "./gradlew", "build")
	if err != nil {
		return nil, err
	}

	s, err := os.Open(projectDir + "/app/build/outputs/apk")
	if err != nil {
		return nil, err
	}
	defer s.Close()

	ls, err := s.Readdir(-1)
	if err != nil {
		return nil, err
	}

	paths := make([]string, len(ls))
	for i, l := range ls {
		paths[i] = projectDir + "/app/build/outputs/apk/" + l.Name()
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
