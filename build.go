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
		projects, err := storage.ProjectsList()
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
func update(project string) error {
	sourceDir := storage.SourcePath(project)

	g := git{sourceDir: sourceDir}

	// Update the source.
	err := g.fetch()
	if err != nil {
		return err
	}

	branches, err := g.branches()
	if err != nil {
		return err
	}

	for _, branch := range branches {
		latestSourceID := branch.desc
		latestBuildID, err := storage.LatestVersion(project, branch.name)
		if err != nil {
			return err
		}
		// If no updates, skip.
		if latestBuildID == latestSourceID {
			continue
		}
		log.Printf("%s: building %s %s", project, branch.name, branch.desc)

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

		logger, err := storage.BuildLogger(project, branch.name, latestSourceID)
		if err != nil {
			return err
		}
		files, err := runBuilds(sourceDir, logger)
		logger.Close()

		if err == nil {
			err = storage.SaveBuilds(project, branch.name, latestSourceID, files)
			if err == nil {
				log.Printf("%s: saved %v", project, files)
			}
		} else {
			log.Printf("%s: build failed: %v", project, err)
		}

		err = storage.SetLatestBuildID(project, branch.name, latestSourceID)
		if err != nil {
			log.Println(err)
		}
	}

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
func runBuilds(sourceDir string, logger io.Writer) ([]string, error) {
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
			files, err := builder.Build(logger, append(os.Environ(), toEnvList(versionCfg.Env)...))
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
