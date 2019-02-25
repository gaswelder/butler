package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/gaswelder/butler/builders"
)

func (p *project) update() error {
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
		err := g.checkout(branch)
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
		latestBuildID, err := p.latestBuildID(branch)
		if err != nil {
			return err
		}
		log.Printf("branch: %s, latest source: %s, latest build: %s", branch, latestSourceID, latestBuildID)
		if latestBuildID == latestSourceID {
			log.Print("Nothing new, skipping")
			return nil
		}

		// One source root may have multiple buildable things (monorepo).
		bs, err := scanProject(sourceDir)
		if err != nil {
			return err
		}
		for _, builder := range bs {
			log.Printf("%s -> %s", builder.Dirname(), builder.Name())
		}

		for _, builder := range bs {
			files, err := builder.Build()
			if err != nil {
				return err
			}

			// Publish the builds
			pubDir := fmt.Sprintf("%s/builds/%s/%s", p.dir, branch, latestSourceID)
			err = copyFiles(files, pubDir)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		// Update the latest mark.
		err = p.setLatestBuildID(branch, latestSourceID)
		if err != nil {
			return err
		}
	}

	return nil
}

func scanProject(sourceDir string) ([]builders.Builder, error) {
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

func copyFiles(paths []string, to string) error {
	if !exists(to) {
		err := os.MkdirAll(to, 0777)
		if err != nil {
			return err
		}
	}
	for _, f := range paths {
		err := run(".", "cp", f, to+"/"+path.Base(f))
		if err != nil {
			return err
		}
	}
	return nil
}
