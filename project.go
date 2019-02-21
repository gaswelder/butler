package main

// env: {
//     defaults: {
//         APP_ID: "io.iterators.supercryptMobile"
//     },
//     dev: {
//         APP_NAME: "Rewarders Dev"
//     },
//     stage: {
//         APP_NAME: "Rewarders Stage",
//         MixpanelAPIToken: "306730f6c4c0fd02bd68705a6c965737"
//     }
// }

// echo "Getting appropriate tag info (latest in the current branch or just current tag)"
// 	TAG=$(git describe --tags --abbrev=0)
// 	echo "Tag=${TAG}"
//     TAG_VERSION=$(echo ${TAG} | cut -d "-" -f 1)
// APP_VERSION // 1.1.0
// APP_VERSION_DASH // 7
// APP_VERSION_ID // 1.1.0-135-f10dsfsd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type build struct {
	path string
}

func (b *build) url() string {
	return b.path
}

type project struct {
	name string
	dir  string
}

func listProjects() ([]*project, error) {
	f, err := os.Open("projects")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ls, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	result := make([]*project, len(ls))
	for i, l := range ls {
		result[i] = &project{
			name: l.Name(),
			dir:  fmt.Sprintf("projects/%s", l.Name()),
		}
	}
	return result, nil
}

func getProject(name string) *project {
	dir := "projects/" + name
	if !exists(dir) {
		return nil
	}
	return &project{
		name: name,
		dir:  dir,
	}
}

func (p *project) builds() ([]*build, error) {
	f, err := os.Open(p.dir + "/builds")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ls, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	result := make([]*build, len(ls))
	for i, l := range ls {
		result[i] = &build{
			path: p.dir + "/builds/" + l.Name(),
		}
	}
	return result, nil
}

func (p *project) update() error {
	log.Printf("updating %v", p.name)
	if p.name == "bogan" {
		return nil
	}
	var err error

	sourceDir := fmt.Sprintf("%s/src", p.dir)

	// Update the source.
	err = run(sourceDir, "git", "checkout", "master")
	if err != nil {
		return err
	}
	// err = run(sourceDir, "git", "pull")
	// if err != nil {
	// 	return err
	// }

	// If nothing new, stop.
	latestSourceID, err := runOut(sourceDir, "git", "describe")
	if err != nil {
		return err
	}
	latestBuildID, err := p.latestBuildID()
	if err != nil {
		return err
	}
	log.Printf("latest: source %v, build %v", latestSourceID, latestBuildID)
	if latestBuildID == latestSourceID {
		log.Print("Nothing new, skipping")
		return nil
	}

	// Run the build
	outfiles, err := buildReactNative(sourceDir)
	if err != nil {
		return err
	}

	// Update the website.
	err = copyFiles(outfiles, p.dir+"/builds/"+latestSourceID)
	if err != nil {
		return err
	}

	// Update the latest mark.
	err = p.setLatestBuildID(latestSourceID)
	if err != nil {
		return err
	}

	return nil
}

func copyFiles(paths []string, to string) error {
	if !exists(to) {
		err := os.Mkdir(to, 0777)
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

func (p *project) latestBuildID() (string, error) {
	path := fmt.Sprintf("%s/latest", p.dir)
	b, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return "(no build)", nil
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (p *project) setLatestBuildID(id string) error {
	path := fmt.Sprintf("%s/latest", p.dir)
	return ioutil.WriteFile(path, []byte(id), 0666)
}
