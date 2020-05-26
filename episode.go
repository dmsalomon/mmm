package main

import (
	"errors"
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"
)

var (
	SubExt = []string{"srt", "sub", "ass"}
	VidExt = []string{"mp4", "mkv", "avi"}
	AllExt = append(SubExt, VidExt...)

	Regexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(.+?)s(\d\d)[\Wx]*e(\d+).*?\.(\w{3})$`),
		regexp.MustCompile(`(?i)^(.+?)\.(\d+)x(\d+)\..*?\.(\w{3})$`),
		regexp.MustCompile(`(?i)^(.+?)\.(\d)(\d\d+)\..*?\.(\w{3})$`),
	}

	JunkEndRegex = regexp.MustCompile(`[\W_\.]+$`)
	JunkRegex    = regexp.MustCompile(`[\W_\.]`)
	YearRegex    = regexp.MustCompile(`\d{4}`)
	SampleRegex  = regexp.MustCompile(`(?i)SAMPLE`)

	SampleErr = errors.New("File is a sample")
	MatchErr  = errors.New("Filename does not follow recognized pattern")
	ExtErr    = errors.New("Invalid extension")
)

type Episode struct {
	path    string
	file    string
	show    string
	ext     string
	dest    string
	season  uint
	episode uint
}

func (e *Episode) String() string {
	return fmt.Sprintf("%s.s%02de%02d", e.show, e.season, e.episode)
}

func NewEpisode(_path string) (*Episode, error) {
	var match []int

	file := path.Base(_path)

	if SampleRegex.MatchString(file) {
		return nil, SampleErr
	}

	parsed := false
	for _, re := range Regexes {
		match = re.FindStringSubmatchIndex(file)
		if len(match) == 0 {
			continue
		}

		parsed = true
		break
	}
	if !parsed {
		return nil, MatchErr
	}

	show := file[match[2]:match[3]]
	show = JunkEndRegex.ReplaceAllLiteralString(show, "")
	show = JunkRegex.ReplaceAllLiteralString(show, "-")

	season, err := strconv.Atoi(file[match[4]:match[5]])
	if err != nil {
		return nil, err
	}

	episode, err := strconv.Atoi(file[match[6]:match[7]])
	if err != nil {
		return nil, err
	}

	ext := file[match[8]:match[9]]
	if !contains(AllExt, ext) {
		return nil, ExtErr
	}

	e := &Episode{
		path:    _path,
		file:    file,
		show:    show,
		season:  uint(season),
		episode: uint(episode),
	}
	return e, nil
}

func main() {
	path := "/media/trans/Game.of.Thrones.S03e41.720p.mkv"
	e, err := NewEpisode(path)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("---------------------------")

	if e != nil {
		fmt.Println(e.show)
		fmt.Println(e.season)
		fmt.Println(e.episode)
		fmt.Println(e)
	} else {
		fmt.Println("cannot parse")
	}
}
