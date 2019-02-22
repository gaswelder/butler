package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/gaswelder/butler/builders"
)

func (p *project) update() error {
	if p.name == "bogan" {
		return nil
	}
	sourceDir := fmt.Sprintf("%s/src", p.dir)

	g := git{sourceDir: sourceDir}

	// Update the source.
	err := g.pull()
	if err != nil {
		return err
	}

	branches, err := g.branches()
	if err != nil {
		return err
	}
	log.Println(branches)

	for _, branch := range branches {
		err := g.checkout(branch)
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

		// Find a builder for this kind of project.
		builder, err := findBuilder(sourceDir)
		if err != nil {
			log.Println(err)
			continue
		}

		files, err := builder.Build()
		if err != nil {
			log.Println(err)
			continue
		}

		// Publish the builds
		pubDir := fmt.Sprintf("%s/builds/%s/%s", p.dir, branch, latestSourceID)
		err = copyFiles(files, pubDir)
		if err != nil {
			log.Println(err)
			continue
		}

		// Update the latest mark.
		err = p.setLatestBuildID(branch, latestSourceID)
		if err != nil {
			return err
		}
	}

	return nil
}

func findBuilder(sourceDir string) (builders.Builder, error) {
	// Just return a React Native builder for now.
	return builders.ReactNative(sourceDir), nil
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
