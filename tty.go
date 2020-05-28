package main

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

var isTerm = terminal.IsTerminal(int(os.Stdout.Fd()))

func blue() string {
	return bold("34")
}

func white() string {
	return bold("39")
}

func red() string {
	return underline("31")
}

func yellow() string {
	return underline("33")
}

func reset() string {
	return escape("0")
}

func em() string {
	return underline("39")
}

func green() string {
	return bold("32")
}

func gray() string {
	return bold("30")
}

func highlight() string {
	return bold("39")
}

func color(n string) string {
	return escape(fmt.Sprintf("0;%s", n))
}

func bold(n string) string {
	return escape(fmt.Sprintf("1;%s", n))
}

func underline(n string) string {
	return escape(fmt.Sprintf("4;%s", n))
}

func escape(seq string) string {
	if isTerm {
		return fmt.Sprintf("\033[%sm", seq)
	} else {
		return ""
	}
}

func oh1(title string) {
	fmt.Printf("%s==>%s %s%s\n", blue(), white(), title, reset())
}

func oh1f(format string, args ...interface{}) {
	fmt.Printf("%s==>%s ", blue(), white())
	fmt.Printf(format, args...)
	fmt.Printf("%s\n", reset())
}

func ohai(title string, args ...interface{}) {
	fmt.Printf("%s==>%s %s%s\n", blue(), white(), title, reset())
	for _, v := range args {
		fmt.Println(v)
	}
}

func opoof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%sWarning:%s ", yellow(), reset())
	fmt.Fprintf(os.Stderr, format, args...)
}

func opoo(err interface{}) {
	opoof("%v\n", err)
}

func onoef(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%sError:%s ", red(), reset())
	fmt.Fprintf(os.Stderr, format, args...)
}

func onoe(err interface{}) {
	onoef("%v\n", err)
}

func odie(err interface{}) {
	onoe(err)
	os.Exit(1)
}
