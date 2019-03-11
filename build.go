package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gaswelder/butler/builders"
	"github.com/gaswelder/butler/storage"
)

// trackUpdates continuously updates the projects directory
// and makes new builds.
func trackUpdates() {
	for {
		projects, err := storage.Projects()
		if err != nil {
			// If something went wrong while trying to get the list of
			// projects, wait a little before trying again.
			log.Printf("failed to get projects list: %v", err)
			time.Sleep(60 * time.Second)
			continue
		}

		for _, project := range projects {
			err := update(project)
			if err != nil {
				log.Printf("failed to update %s: %v", project, err)
				time.Sleep(60 * time.Second)
				continue
			}
		}

		// Sleep a while and repeat the whole thing again.
		time.Sleep(10 * time.Second)
	}
}

// update updates all builds for the given project.
func update(project storage.Project) error {
	sourceDir := storage.SourcePath(project.Name)

	g := git{sourceDir: sourceDir}

	// Update the source.
	err := g.fetch()
	if err != nil {
		return err
	}

	// Build tips of all branches, but also build the latest clean tag.
	branches, err := g.branches()
	if err != nil {
		return err
	}
	tags, err := g.tags()
	if err != nil {
		return fmt.Errorf("couldn't get tags list: %v", err)
	}

	if len(tags) > 0 {
		tag := tags[len(tags)-1]
		if !storage.Has(project.Name, storage.ReleasesDirectory, tag) {
			log.Printf("%s: building %s", project.Name, tag)
			// Builds on previous branches might change the source tree, so
			// we have to do a reset.
			err = g.discard()
			if err != nil {
				return err
			}
			err = g.checkout(tag)
			if err != nil {
				return err
			}
			err = build(project, storage.ReleasesDirectory, tag)
			if err != nil {
				log.Printf("tag build failed: %v", err)
			}
		}
	}

	for _, branch := range branches {
		if !storage.Has(project.Name, branch.name, branch.desc) {
			log.Printf("%s: building %s %s", project.Name, branch, branch.desc)
			// Builds on previous branches might change the source tree, so
			// we have to do a reset.
			err = g.discard()
			if err != nil {
				return err
			}
			err = g.checkout(branch.name)
			if err != nil {
				return err
			}

			err = g.pull()
			if err != nil {
				return err
			}
			err = build(project, branch.name, branch.desc)
			if err != nil {
				log.Printf("branch build failed: %v", err)
			}
		}
	}

	return nil
}

func build(project storage.Project, directory, version string) error {
	var err error

	sourceDir := storage.SourcePath(project.Name)
	logger, err := storage.BuildLogger(project.Name, directory, version)
	if err != nil {
		return err
	}
	files, err := runBuilds(sourceDir, logger, project.Env)
	logger.Close()
	if err != nil {
		return fmt.Errorf("build failed: %v", err)
	}

	err = storage.SaveBuilds(project.Name, directory, version, files)
	if err != nil {
		return fmt.Errorf("failed to save builds: %v", err)
	}
	log.Printf("%s: saved %v", project, files)
	return nil
}

type versionConfig struct {
	Env map[string]string `json:"env"`
}

type sourceConfig struct {
	Versions map[string]versionConfig `json:"versions"`
}

func config(sourceDir string) (*sourceConfig, error) {
	cfg := &sourceConfig{
		Versions: map[string]versionConfig{
			"dev": {
				Env: map[string]string{
					"BUTLER_ENV": "dev",
				},
			},
		},
	}

	// Read butler.json. If no such file, return the default config.
	data, err := ioutil.ReadFile(sourceDir + "/butler.json")
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse butler.json: %v", err)
	}
	return cfg, nil
}

func toEnvList(vars map[string]string) []string {
	list := make([]string, 0)
	for k, v := range vars {
		list = append(list, k+"="+v)
	}
	return list
}

// runBuilds builds everything in the given source directory and returns
// a list of build outputs.
func runBuilds(sourceDir string, logger io.Writer, env []string) ([]string, error) {
	// Get builders for this project.
	bs, err := detectBuilders(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("could not get project builders: %s", err.Error())
	}
	if len(bs) == 0 {
		return nil, fmt.Errorf("no builders detected")
	}
	for _, builder := range bs {
		log.Printf("%s -> %s", builder.Dirname(), builder.Name())
	}

	cfg, err := config(sourceDir)
	if err != nil {
		return nil, err
	}

	allFiles := make([]string, 0)
	for _, builder := range bs {
		for envName, versionCfg := range cfg.Versions {
			// Combine all environment variables in one list.
			versionEnv := append(os.Environ(), env...)
			versionEnv = append(versionEnv, toEnvList(versionCfg.Env)...)

			files, err := builder.Build(logger, versionEnv)
			if err != nil {
				return nil, err
			}
			// stash the files somewhere before they get deleted by the next build.
			s, err := storage.Stash(files, envName)
			if err != nil {
				return nil, err
			}
			allFiles = append(allFiles, s...)
		}
	}
	return allFiles, nil
}

// detectBuilders returns a list of builders needed for the given source directory.
func detectBuilders(sourceDir string) ([]builders.Builder, error) {
	// Check if this is a one-project source.
	builder, err := builders.Find(sourceDir)
	if err != nil {
		return nil, err
	}
	if builder != nil {
		return []builders.Builder{builder}, nil
	}

	// If not, then we might have multiple projects.
	dir, err := os.Open(sourceDir)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	ls, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	bs := make([]builders.Builder, 0)
	for _, f := range ls {
		if !f.IsDir() {
			continue
		}
		if f.Name()[0] == '.' {
			continue
		}
		builder, err := builders.Find(sourceDir + "/" + f.Name())
		if err != nil {
			return nil, err
		}
		if builder != nil {
			bs = append(bs, builder)
		}
	}
	return bs, nil
}
