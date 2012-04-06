package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	startPrefix = "=== RUN "
	passPrefix = "--- PASS: "
	failPrefix = "--- FAIL: "
)

var endRegexp *regexp.Regexp

type Test struct {
	Name string
	Time time.Duration
	Message string
	Failed bool
}

func parseEnd(prefix, line string) (string, time.Duration, error) {
	matches := endRegexp.FindStringSubmatch(line[len(prefix):])

	if len(matches) == 0 {
		return "", 0, fmt.Errorf("can't parse %s", line)
	}

	seconds, _ := strconv.ParseFloat(matches[2], 32)
	return matches[1], time.Duration(seconds * float64(time.Second)), nil
}

func parseOutput(rd io.Reader) ([]*Test, error) {
	tests := []*Test{}

	reader := bufio.NewReader(rd)
	var test *Test = nil
	for {
		/* FIXME: Handle isPrefix */
		buf, _, err := reader.ReadLine()

		switch err {
		case io.EOF:
			if test != nil {
				tests = append(tests, test)
			}
			return tests, nil
		case nil:
			;
		default:
			return nil, err
		}

		line := string(buf)
		switch {
		case strings.HasPrefix(line, startPrefix):
			if test != nil {
				tests = append(tests, test)
			}
			test = &Test{Name: line[len(startPrefix):]}
		case strings.HasPrefix(line, failPrefix):
			if test == nil {
				return nil, fmt.Errorf("fail not inside test")
			}
			test.Failed = true
			name, time, err := parseEnd(failPrefix, line)
			if err != nil {
				return nil, err
			}
			if name != test.Name {
				return nil, fmt.Errorf("wrong test end (%s!=%s)", name, test.Name)
			}
			test.Time = time
		case strings.HasPrefix(line, passPrefix):
			if test == nil {
				return nil, fmt.Errorf("pass not inside test")
			}
			test.Failed = false
			name, time, err := parseEnd(passPrefix, line)
			if err != nil {
				return nil, err
			}
			if name != test.Name {
				return nil, fmt.Errorf("wrong test end (%s!=%s)", name, test.Name)
			}
			test.Time = time
		default:
			if test != nil {
				test.Message += line + "\n"
			}
		}
	}

	return tests, nil
}

func main() {
	var input io.Reader
	var err error

	flag.Parse()
	if len(flag.Args()) > 0 {
		input, err = os.Open(flag.Args()[0])
		if err != nil {
			fmt.Printf("error: can't open %s: %s", flag.Args()[0], err)
			os.Exit(1)
		}
	} else {
		input = os.Stdin
	}

	endRegexp = regexp.MustCompile(`([^ ]+) \((\d+\.\d+)`)
	tests, err := parseOutput(input)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("%v\n", tests)

	for _, test := range tests {
		fmt.Println(test.Name)
	}
}
