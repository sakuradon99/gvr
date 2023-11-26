package main

import (
	"flag"
	"fmt"
	"gvr/internal/git"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var majorFlag = flag.Bool("m", false, "increase major version")
var minorFlag = flag.Bool("n", false, "increase minor version")

func main() {
	flag.Parse()
	err := handleError(run())
	if err != nil {
		panic(err)
	}
}

func run() error {
	status, err := git.Status()
	if err != nil {
		return err
	}

	if !strings.Contains(status, "nothing to commit, working tree clean") {
		fmt.Println("has uncommitted changes")
		return nil
	}

	var versions []version

	var latestVersion version
	if len(versions) > 0 {
		sort.SliceIsSorted(versions, func(i, j int) bool {
			if versions[i].major != versions[j].major {
				return versions[i].major > versions[j].major
			}
			if versions[i].minor != versions[j].minor {
				return versions[i].minor > versions[j].minor
			}
			return versions[i].patch > versions[j].patch
		})
		latestVersion = versions[0]
	}

	newVersion := latestVersion
	if *majorFlag {
		newVersion.major++
		newVersion.minor = 0
		newVersion.patch = 0
	} else if *minorFlag {
		newVersion.minor++
		newVersion.patch = 0
	} else {
		newVersion.patch++
	}

	fmt.Printf("new version: v%d.%d.%d (Y/n): ", newVersion.major, newVersion.minor, newVersion.patch)
	var userInput string
	_, err = fmt.Scanln(&userInput)
	if err != nil {
		return err
	}
	if userInput != "Y" {
		return nil
	}

	return nil
}

func handleError(err error) error {

	return err
}

var versionRegexp = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

type version struct {
	major int
	minor int
	patch int
}

func (v version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func parseVersion(tag string) (version, bool) {
	matches := versionRegexp.FindStringSubmatch(tag)
	if len(matches) != 4 {
		return version{}, false
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return version{}, false
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return version{}, false
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return version{}, false
	}

	return version{
		major: major,
		minor: minor,
		patch: patch,
	}, true
}
