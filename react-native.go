package main

import "os"

func buildReactNative(sourceDir string) ([]string, error) {
	var err error
	err = npm(sourceDir)
	if err != nil {
		return nil, err
	}

	err = run(sourceDir+"/android", "./gradlew", "build")
	if err != nil {
		return nil, err
	}

	// ./gradlew clean
	// ./gradlew assemble${BUILD_ENV}Release
	// run(sourceDir, "bash butler.sh")
	s, err := os.Open(sourceDir + "/android/app/build/outputs/apk")
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
		paths[i] = sourceDir + "/android/app/build/outputs/apk/" + l.Name()
	}
	return paths, nil
}
