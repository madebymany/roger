package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var minSpecStr = flag.String("mins", "*", "cron-style minute spec")
var hourSpecStr = flag.String("hours", "*", "cron-style hour spec")
var inShell = flag.Bool("shell", false, "Run command in a shell")

type timeSpec struct {
	every     int
	instances []int
}

var explicitTimeSpecPattern = regexp.MustCompile(`\A(([\d,])+|\*)(/(\d+))?\z`)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatalln("Command cannot be empty")
	}

	minSpec := parseTimeSpec(*minSpecStr)
	hourSpec := parseTimeSpec(*hourSpecStr)

	var cmd *exec.Cmd
	if *inShell {
		cmd = exec.Command("/bin/sh", "-c", flag.Args()[0])
	} else {
		cmd = exec.Command(flag.Args()[0], flag.Args()[1:]...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var now time.Time
	for {
		now = time.Now()

		if now.Second() == 0 && minSpec.matches(now.Minute()) &&
			hourSpec.matches(now.Hour()) {

			// TODO: catch error and report somehow
			cmd.Run()
		}

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

func (self timeSpec) matches(moment int) bool {
	if (moment % self.every) != 0 {
		return false
	}

	if self.instances == nil {
		return true
	}
	for _, i := range self.instances {
		if moment == i {
			return true
		}
	}
	return false
}
