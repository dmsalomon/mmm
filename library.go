package main

import (
	"github.com/sahilm/fuzzy"
	"io/ioutil"
	"path"
)

var (
	shows = []string{}

	MediaPaths = []string{
		"/media/green/tv",
		"/media/black/tv",
		"/media/pink/tv",
		"/media/pink/Anime",
	}
)

type FuzzyLib []string

func (f FuzzyLib) String(i int) string {
	return path.Base(f[i])
}

func (f FuzzyLib) Len() int {
	return len(f)
}

func init() {
	for _, lib := range MediaPaths {
		files, err := ioutil.ReadDir(lib)

		if err != nil {
			return
		}

		for _, file := range files {
			if file.IsDir() {
				shows = append(shows, path.Join(lib, file.Name()))
			}
		}
	}
}

func LibSearch(show string) string {
	matches := fuzzy.FindFrom(show, FuzzyLib(shows))
	for _, match := range matches {
		return match.Str
	}
	return ""
}
