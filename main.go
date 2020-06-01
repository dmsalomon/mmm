package main

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"
)

var (
	logger *zap.SugaredLogger

	SpyDir = "/media/pink/transmission/completed"

	Replace = false
)

const (
	MaxUint = ^uint(0)
)

func init() {
	cfg := zap.NewProductionConfig()
	if isTerm {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level.SetLevel(zapcore.InfoLevel)
	}
	rawlog, err := cfg.Build()
	if err != nil {
		odie(err)
	}

	logger = rawlog.Sugar()
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
	dirpath, err := filepath.Abs(dir)
	if err != nil {
		onoef("%v: %v\n", dir, err)
		return false
	}

	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		onoef("%v: %v\n", dir, err)
		return false
	}

	ok := true
	for _, fi := range files {
		var good bool

		fn := fi.Name()
		if len(fn) > 0 && fn[0] == '.' {
			// skip dotfiles
			continue
		}

		path := path.Join(dirpath, fn)
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
	if Replace {
		e.tier = MaxUint
	}
	err = e.Install()
	if err != nil {
		onoef("%v: %v\n", e, err)
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

func spy(dir string) {
	reloadLib := make(chan struct{}, 1)
	for _, path := range MediaPaths {
		go func(path string) {
			inotifywait(path, "create,move,delete", false, func(event string) {
				select {
				case reloadLib <- struct{}{}:
				default:
				}
			})
		}(path)
	}
	go func() {
		for {
			<-reloadLib
			logger.Info("reloading library")
			to := time.After(1 * time.Second)
			loadLib()
			<-to
		}
	}()

	events := make(chan string, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP)

	go func() {
		for {
			<-sigs
			logger.Info("SIGHUP")
			events <- "SIGHUP"
		}
	}()

	go func() {
		inotifywait(dir, "create,moved_to", true, func(event string) {
			events <- event
		})
	}()

	select {
	case events <- "INIT":
	default:
	}

	for {
		logger.Debug(<-events)
		installDir(dir)
	}
}

func Arg(i int) string {
	if i >= len(os.Args) {
		return ""
	} else {
		return os.Args[i]
	}
}

func Args(i, j int) []string {
	a := os.Args
	l := len(a)
	if i >= l || j >= l {
		return []string{}
	}
	if j < 0 {
		j += (l + 1)
	}
	return a[i:j]
}

func Usage() {
	out := os.Stderr
	fmt.Fprintf(out, "Usage: %s [OPTION] [CMD]\n", os.Args[0])
	fmt.Fprintln(out, "CMD: cd, install, replace, spy")
}

func main() {
	defer logger.Sync()

	ok := true

	switch Arg(1) {
	case "cd":
		ok = cd(Arg(2))
	case "replace":
		Replace = true
		fallthrough
	case "install":
		ok = installList(Args(2, -1))
	case "spy":
		spy(SpyDir)
	default:
		Usage()
	}

	if !ok {
		os.Exit(1)
	}
}
