package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gaswelder/butler/builders"
)

type build struct {
	path string
}

func (b *build) url() string {
	return b.path
}

// trackUpdates continuously updates the projects directory
// and makes new builds.
func trackUpdates() {
	for {
		projects, err := listProjects()
		if err != nil {
			// If something went wrong while trying to get the list of
			// projects, wait a little before trying again.
			log.Printf("failed to get projects list: %v", err)
			time.Sleep(60 * time.Second)
			continue
		}

		for _, project := range projects {
			log.Printf("Project: %s", project.name)
			err := update(project)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Sleep a while and repeat the whole thing again.
		time.Sleep(10 * time.Second)
	}
}

// update updates all builds for the given project.
func update(p *project) error {
	sourceDir := fmt.Sprintf("%s/src", p.dir)

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
	log.Println("branches:", branches)

	for _, branch := range branches {
		// Builds on previous branches might change the source tree, so
		// we have to do a reset.
		err := g.discard()
		if err != nil {
			return err
		}
		err = g.checkout(branch)
		if err != nil {
			return err
		}
		err = g.pull()
		if err != nil {
			return err
		}
		latestSourceID, err := g.describe()
		if err != nil {
			return err
		}
		latestBuildID, err := latestBuildID(p.name, branch)
		if err != nil {
			return err
		}
		log.Printf("branch: %s, latest source: %s, latest build: %s", branch, latestSourceID, latestBuildID)
		if latestBuildID == latestSourceID {
			log.Print("Nothing new, skipping")
			continue
		}

		logger, err := buildLogger(p.name, branch, latestSourceID)
		if err != nil {
			return err
		}
		files, err := runBuilds(sourceDir, logger)
		logger.Close()

		if err == nil {
			err = publish(p.dir, branch, latestSourceID, files)
		}
		if err != nil {
			log.Println(err)
		}

		// Update the latest mark.
		err = setLatestBuildID(p.name, branch, latestSourceID)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	return nil
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

	allFiles := make([]string, 0)
	for _, builder := range bs {
		files, err := builder.Build(logger)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, files...)
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
