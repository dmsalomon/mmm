package main

import (
	"github.com/sahilm/fuzzy"
	"io/ioutil"
	"path"
)

var (
	Shows = []string{}
	Paths = map[string]string{}

	MediaPaths = []string{
		"/media/green/tv",
		"/media/black/tv",
		"/media/pink/tv",
		"/media/pink/Anime",
	}
)

func init() {
	for _, lib := range MediaPaths {
		files, err := ioutil.ReadDir(lib)

		if err != nil {
			return
		}

		for _, file := range files {
			if file.IsDir() {
				fn := file.Name()
				Shows = append(Shows, fn)
				Paths[fn] = path.Join(lib, fn)

			}
		}
	}
}

func LibSearch(show string) string {
	matches := fuzzy.Find(show, Shows)
	for _, match := range matches {
		return match.Str
	}
	return ""
}
