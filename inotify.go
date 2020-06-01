package main

import (
	"bufio"
	"os/exec"
)

func inotifywait(path string, eventlist string, recursive bool, handler func(string)) {
	var cmd *exec.Cmd
	reload := false
begin:
	events := make(chan string)
	if recursive {
		cmd = exec.Command("inotifywait", "-m", "-q", "-r", "-e", eventlist, path)
	} else {
		cmd = exec.Command("inotifywait", "-m", "-q", "-e", eventlist, path)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		odie(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		odie(err)
	}

	if err := cmd.Start(); err != nil {
		odie(err)
	}
	if reload {
		logger.Info("exec inotifywait")
	}

	go func() {
		s := bufio.NewScanner(stdout)
		for s.Scan() {
			events <- s.Text()
		}
	}()
	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			events <- s.Text()
		}
	}()

	reset := make(chan struct{})
	go func() {
		cmd.Wait()
		reset <- struct{}{}
	}()

	for {
		select {
		case e := <-events:
			handler(e)
		case <-reset:
			reload = true
			goto begin
		}

	}
}
