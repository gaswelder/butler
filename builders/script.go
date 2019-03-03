package builders

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

// ScriptBuilder is a builder calling a custom script.
type ScriptBuilder struct {
	projectDir string
}

// Script returns a builder for a React Native project.
func Script(projectDir string) Builder {
	return &ScriptBuilder{
		projectDir: projectDir,
	}
}

// Build builds the project.
func (b *ScriptBuilder) Build(output io.Writer, envVars []string) ([]string, error) {
	err := os.MkdirAll("./tmp/scriptbuilder", 0777)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	tmpDir, err := ioutil.TempDir(cwd+"/tmp/scriptbuilder", "")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("./butler.sh", tmpDir)
	cmd.Dir = b.projectDir
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.Env = envVars

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	return ls(tmpDir)
}

// Name returns the builder's name.
func (b *ScriptBuilder) Name() string {
	return "Script"
}

// Dirname returns the builder's project path.
func (b *ScriptBuilder) Dirname() string {
	return b.projectDir
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
		if l.IsDir() {
			continue
		}
		result[i] = dir + "/" + l.Name()
	}
	return result, nil
}
