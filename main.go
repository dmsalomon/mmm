package main

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

var (
	logger *zap.SugaredLogger

	SpyDir = "/media/pink/transmission/completed"

	// flags
	DryRun bool
)

func init() {
	rawlog, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger = rawlog.Sugar()
}

func init() {
	flag.BoolVar(&DryRun, "dry", false, "Dry Run")
}

func shit() {
	fmt.Println("---------------------------")

	fmt.Printf("%v = %v\n", "dry", DryRun)

	path := "/tmp/The.Americans.S06E01.AMZN.mkv"
	e, err := NewEpisode(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	if e != nil {
		fmt.Println(e.show)
		fmt.Println(e.season)
		fmt.Println(e.episode)
		fmt.Println("Release:", e.tier)
		fmt.Println(e)
	} else {
		fmt.Println("cannot parse")
	}

	fmt.Println("trimmed:", trim(e.show))

	m := libSearch(e.show)
	fmt.Println(m)
	fmt.Println(Paths[m])

	_, episodes, err := loadSeason(e.show, e.season)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("len(episodes):", len(episodes))

	if episodes != nil {
		for _, ep := range episodes {
			fmt.Println(ep.path)
		}
	}

	err = e.Install()
	if err != nil {
		fmt.Println(err)
	}
}

func cd(show string) bool {
	if show == "" {
		return false
	}

	showpath := showPath(show)

	if showpath == "" {
		fmt.Printf("%s\n", ".")
		return false
	} else {
		fmt.Printf("%s\n", showpath)
		return true
	}
}

func installDir(dir string) bool {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		onoef("%v: %v\n", dir, err)
		return false
	}

	ok := true
	for _, fi := range files {
		var good bool
		path := path.Join(dir, fi.Name())
		if fi.IsDir() {
			good = installDir(path)
		} else {
			good = install(path)
		}
		if !good {
			ok = false
		}
	}
	return ok
}

func install(file string) bool {
	e, err := NewEpisode(file)
	if err != nil {
		onoef("%v: %v\n", file, err)
		return false
	}
	err = e.Install()
	if err != nil {
		onoef("%v: %v\n", file, err)
		return false
	}
	return true
}

func installList(files []string) bool {
	ok := true
	for _, file := range files {
		path, err := filepath.Abs(file)
		if err != nil {
			onoef("%v: %v\n", file, err)
			ok = false
			continue
		}

		fi, err := os.Stat(path)
		if err != nil {
			onoef("%v: %v\n", file, err)
			ok = false
			continue
		}

		var good bool
		if fi.IsDir() {
			good = installDir(file)
		} else {
			good = install(file)
		}

		if !good {
			ok = false
		}
	}
	return ok
}

func main() {
	defer logger.Sync()
	flag.Parse()

	switch flag.Arg(0) {
	case "cd":
		cd(flag.Arg(1))
	case "install":
		installList(flag.Args()[1:])
	default:
		fmt.Println("damn")
	}
}
