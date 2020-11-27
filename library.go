package main

import (
	"errors"
	"fmt"
	"github.com/sahilm/fuzzy"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	libLock      sync.Mutex
	Shows        []string
	ShowsTrimmed []string
	Paths        map[string]string

	MediaPaths = []string{
		"/media/green/tv",
		"/media/black/tv",
		"/media/pink/tv",
		"/media/pink/Anime",
	}

	CachePath = "/media/pink/mmmcache"

	ReleaseTier = map[string]uint{
		"PROPER": 1,
		"Atmos":  1,
		"AMZN":   2,
	}

	ResTier = map[string]uint{
		"720p":  1,
		"1080p": 2,
	}

	// ordered from highest priority to lowest
	Tiers = []map[string]uint{ResTier, ReleaseTier}

	ShowNotFoundErr   = errors.New("show not found")
	SeasonNotFoundErr = errors.New("season not found")
	DupEpisodeErr     = errors.New("episode already in library")

	ExtRegex = regexp.MustCompile(`\.\w{3}$`)
)

type EpisodeList []*Episode

func (e EpisodeList) Len() int           { return len(e) }
func (e EpisodeList) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e EpisodeList) Less(i, j int) bool { return e[i].episode < e[j].episode }

func init() {
	loadLib()
}

func trim(s string) string {
	up := strings.ToLower(s)
	post := strings.TrimPrefix(up, "the")
	return strings.Replace(post, "-", "", -1)
}

func eqbase(a, b string) bool {
	a = ExtRegex.ReplaceAllLiteralString(a, "")
	b = ExtRegex.ReplaceAllLiteralString(b, "")

	l := min(len(a), len(b))

	return a[:l] == b[:l]
}

func libSearch(show string) string {
	matches := fuzzy.Find(trim(show), ShowsTrimmed)
	for _, match := range matches {
		return Shows[match.Index]
	}
	return ""
}

func showPath(show string) string {
	show = libSearch(show)
	return Paths[show]
}

func loadLib() {
	libLock.Lock()
	defer libLock.Unlock()

	Shows = []string{}
	ShowsTrimmed = []string{}
	Paths = map[string]string{}

	for _, lib := range MediaPaths {
		files, err := ioutil.ReadDir(lib)

		if err != nil {
			return
		}

		for _, file := range files {
			if file.IsDir() {
				fn := file.Name()
				Shows = append(Shows, fn)
				ShowsTrimmed = append(ShowsTrimmed, trim(fn))
				Paths[fn] = path.Join(lib, fn)
			}
		}
	}
}

func loadSeason2(seasonpath string) ([]*Episode, error) {
	files, err := ioutil.ReadDir(seasonpath)
	if err != nil {
		return nil, err
	}

	episodes := []*Episode{}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fn := file.Name()
		path := path.Join(seasonpath, fn)
		e, err := NewEpisode(path)

		if err == ExtErr {
			continue
		} else if err != nil {
			logger.Debug(path, err)
			continue
		}

		episodes = append(episodes, e)
	}

	return episodes, nil
}

func loadSeason(show string, season uint) (string, []*Episode, error) {
	showpath := showPath(show)
	if showpath == "" {
		return "", nil, ShowNotFoundErr
	}

	seasondir := fmt.Sprintf("Season%d", season)
	seasonpath := path.Join(showpath, seasondir)
	seasoninfo, err := loadSeason2(seasonpath)

	if err != nil {
		if os.IsNotExist(err) {
			err = SeasonNotFoundErr
		}
		return seasonpath, nil, err
	}
	if seasoninfo == nil {
		return seasonpath, nil, SeasonNotFoundErr
	}
	return seasonpath, seasoninfo, nil
}

func loadShow(show string) (string, map[uint][]*Episode) {
	showpath := showPath(show)
	if showpath == "" {
		return "", nil
	}

	seasons, err := ioutil.ReadDir(showpath)
	if err != nil {
		return showpath, nil
	}

	showinfo := map[uint][]*Episode{}

	for _, sdirinfo := range seasons {
		if !sdirinfo.IsDir() {
			continue
		}

		seasondir := sdirinfo.Name()

		var nseason uint

		if strings.ToLower(seasondir) == "specials" {
			nseason = 0
		} else if strings.HasPrefix(seasondir, "Season") {
			season := strings.TrimPrefix(seasondir, "Season")
			nseasond, err := strconv.Atoi(season)
			if err != nil {
				logger.Error(err)
				continue
			}
			nseason = uint(nseasond)
		}

		seasonpath := path.Join(showpath, seasondir)
		showinfo[uint(nseason)], err = loadSeason2(seasonpath)
		if err != nil {
			logger.Error(err)
			continue
		}
	}

	return showpath, showinfo
}

func (e *Episode) Install() error {
	libLock.Lock()
	defer libLock.Unlock()

begin:
	var ep *Episode
	seasonpath, seasoninfo, err := loadSeason(e.show, e.season)

	if err == SeasonNotFoundErr {
		if e.season > 0 && e.season < 20 {
			err = os.Mkdir(seasonpath, os.ModePerm)
			if err != nil {
				logger.Info("failed to create directory")
				return SeasonNotFoundErr
			}
			goto begin
		}
		return err
	} else if err != nil {
		return err
	}

	found := false
	for _, ep = range seasoninfo {
		if e.episode == ep.episode {
			found = true
			break
		}
	}

	if found {
		if e.tier > ep.tier {
			destpath := path.Join(seasonpath, e.file)
			fmt.Printf("%s -> %s\n", e.file, seasonpath)
			err = moveFile(e.path, destpath)
			if err != nil {
				return err
			}
			// logger.Infow("backup",
			// 	"file", ep.path,
			// 	"cache", CachePath)
			// cachepath := path.Join(CachePath, ep.file)
			// err := moveFile(ep.path, cachepath)
			// if err != nil {
			// 	return err
			// }
		} else {
			return DupEpisodeErr
		}
	} else {
		destpath := path.Join(seasonpath, e.file)
		fmt.Printf("%s -> %s\n", e.file, seasonpath)
		err = moveFile(e.path, destpath)
		if err != nil {
			return err
		}
	}

	return nil
}
