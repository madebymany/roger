package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"time"
	"regexp"
	"strconv"
	"strings"
)

var minSpecStr = flag.String("mins", "*", "cron-style minute spec")
var hourSpecStr = flag.String("hours", "*", "cron-style hour spec")
var cmdStr = flag.String("cmd", "", "command to run")

type timeSpec struct {
	every int
	instances []int
}

var explicitTimeSpecPattern = regexp.MustCompile(`\A(([\d,])+|\*)(/(\d+))?\z`)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if *cmdStr == "" {
		log.Fatalln("Command cannot be empty")
	}

	minSpec := parseTimeSpec(*minSpecStr)
	hourSpec := parseTimeSpec(*hourSpecStr)

	var found bool
	cmd := exec.Command("/bin/sh", "-c", *cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	for {
		now := time.Now()

		if now.Second() != 0 || (now.Minute() % minSpec.every) != 0 {
			goto WaitContinue
		}

		if minSpec.instances != nil {
			found = false
			for _, min := range minSpec.instances {
				if now.Minute() == min {
					found = true
					break
				}
			}
			if !found {
				goto WaitContinue
			}
		}

		if (now.Hour() % hourSpec.every) != 0 {
			goto WaitContinue
		}

		if hourSpec.instances != nil {
			found = false
			for _, hour := range hourSpec.instances {
				if now.Hour() == hour {
					found = true
					break
				}
			}
			if !found {
				goto WaitContinue
			}
		}

		cmd.Run()

WaitContinue:
		time.Sleep(time.Second)
	}
}

func parseTimeSpec(in string) (out timeSpec) {
	out.every = 1
	var err error

	if in == "*" {
		return
	} else if m := explicitTimeSpecPattern.FindStringSubmatch(in); m != nil {
		if m[1] != "*" {
			instancesStrs := strings.Split(m[1], ",")
			out.instances = make([]int, len(instancesStrs))
			for i, str := range instancesStrs {
				out.instances[i], err = strconv.Atoi(str)
				if err != nil {
					panic(err)
				}
			}
		}

		if m[4] != "" {
			out.every, err = strconv.Atoi(m[4])
			if err != nil {
				panic(err)
			}
		}

		return
	} else {
		log.Fatalf("Invalid spec: '%s'\n", in)
	}
	return
}
