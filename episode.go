package main

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var (
	SubExt = []string{"srt", "sub", "ass"}
	VidExt = []string{"mp4", "mkv", "avi"}
	AllExt = append(SubExt, VidExt...)

	Regexes = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(.+?)s(\d\d)[\Wx]*e(\d+).*?\.(\w{3})$`),
		regexp.MustCompile(`(?i)^(.+?)\.(\d+)x(\d+)\..*?\.(\w{3})$`),
		// regexp.MustCompile(`(?i)^(.+?)\.(\d)(\d\d+)\..*?\.(\w{3})$`),
	}

	JunkEndRegex = regexp.MustCompile(`[\W_\.]+$`)
	JunkRegex    = regexp.MustCompile(`[\W_\.]`)
	YearRegex    = regexp.MustCompile(`\d{4}`)
	SampleRegex  = regexp.MustCompile(`(?i)SAMPLE`)

	SampleErr       = errors.New("File is a sample")
	MatchErr        = errors.New("Filename does not follow recognized pattern")
	ExtErr          = errors.New("Invalid extension")
	FileNotFoundErr = errors.New("File not found")
)

type Episode struct {
	path    string
	file    string
	show    string
	ext     string
	dest    string
	season  uint
	episode uint
	tier    uint
}

func (e *Episode) String() string {
	return fmt.Sprintf("%s.s%02de%02d", e.show, e.season, e.episode)
}

func calcTier(file string) uint {
	tier := uint(0)

	for _, ranking := range Tiers {
		localTier := uint(0)
		for pat, rank := range ranking {
			if strings.Contains(file, pat) {
				localTier = rank
			}
		}
		tier *= uint(1 + len(ranking))
		tier += localTier
	}

	return tier
}

func NewEpisode(_path string) (*Episode, error) {
	var match []int

	if !fexists(_path) {
		return nil, FileNotFoundErr
	}

	file := path.Base(_path)

	if SampleRegex.MatchString(file) {
		return nil, SampleErr
	}

	for _, re := range Regexes {
		match = re.FindStringSubmatchIndex(file)
		if len(match) > 0 {
			goto parsed
		}
	}
	return nil, MatchErr
parsed:

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
	if !contains(VidExt, ext) {
		return nil, ExtErr
	}

	tier := calcTier(file)

	e := &Episode{
		path:    _path,
		file:    file,
		show:    show,
		season:  uint(season),
		episode: uint(episode),
		tier:    tier,
		dest:    "",
	}
	return e, nil
}
