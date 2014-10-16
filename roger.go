package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var minSpecStr = flag.String("mins", "*", "cron-style minute spec")
var hourSpecStr = flag.String("hours", "*", "cron-style hour spec")
var dowSpecStr = flag.String("dow", "*", "cron-style day-of-week spec")
var inShell = flag.Bool("shell", false, "Run command in a shell")
var cmdCwd = flag.String("cwd", "", "Change working directory for command")
var shouldExitFile = flag.String("exitfile", "",
	"File to watch for changes to signal exit")

type timeSpec struct {
	every     int
	instances []int
}

var timeSpecPattern = regexp.MustCompile(`\A(?P<instances>((\d+(-\d+)?,)*(\d+(-\d+)?))+|\*)(/(?P<every>\d+))?\z`)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatalln("Command cannot be empty")
	}

	minSpec := parseTimeSpec(*minSpecStr)
	hourSpec := parseTimeSpec(*hourSpecStr)
	dowSpec := parseTimeSpec(*dowSpecStr)

	if *shouldExitFile != "" {
		log.Println("warning: -exitfile is deprecated. Use 'pkill -x roger' instead.")
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGUSR1)

	var now time.Time

CheckLoop:
	for {
		now = time.Now()

		if now.Second() == 0 &&
			minSpec.matches(now.Minute()) &&
			hourSpec.matches(now.Hour()) &&
			dowSpec.matches(int(now.Weekday())) {

			var cmd *exec.Cmd
			if *inShell {
				cmd = exec.Command("/bin/sh", "-c", flag.Args()[0])
			} else {
				cmd = exec.Command(flag.Args()[0], flag.Args()[1:]...)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = *cmdCwd
			if envCmdDir := os.Getenv("ROGER_CWD"); cmd.Dir == "" {
				cmd.Dir = envCmdDir
			}

			// TODO: catch error and report somehow
			cmd.Run()
		}

		select {
		case _ = <-time.After(time.Second):
			// loop
		case _ = <-signals:
			break CheckLoop
		}
	}
}

func parseTimeSpec(in string) (out timeSpec) {
	out.every = 1

	match := timeSpecPattern.FindStringSubmatch(in)
	if match == nil {
		log.Fatalf("Invalid spec: '%s'\n", in)
		return
	}

	var everyIdx, instancesIdx int
	for i, n := range timeSpecPattern.SubexpNames() {
		switch n {
		case "every":
			everyIdx = i
		case "instances":
			instancesIdx = i
		}
	}

	if match[everyIdx] != "" {
		out.every = mustAtoi(match[everyIdx])
	}

	if match[instancesIdx] != "*" {
		instancesStrs := strings.Split(match[instancesIdx], ",")
		out.instances = make([]int, 0, len(instancesStrs))

		for _, str := range instancesStrs {
			rangeSplit := strings.Split(str, "-")
			switch len(rangeSplit) {
			case 1:
				out.instances = append(out.instances,
					mustAtoi(rangeSplit[0]))
			case 2:
				from, to := mustAtoi(rangeSplit[0]), mustAtoi(rangeSplit[1])
				for i := from; i <= to; i++ {
					out.instances = append(out.instances, i)
				}
			default:
				panic("invalid range")
			}
		}
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

func mustAtoi(s string) (n int) {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return
}
