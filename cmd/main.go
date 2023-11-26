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
	notCommit, err := git.HasChangeNotCommit()
	if err != nil {
		return err
	}
	if notCommit {
		fmt.Println("has uncommitted changes")
		return nil
	}

	notPushed, err := git.HasCommitNotPushed()
	if err != nil {
		return err
	}
	if notPushed {
		fmt.Println("has commit not pushed")
		return nil
	}

	err = git.FetchTags()
	if err != nil {
		return err
	}

	var versions []version
	tagList, err := git.ListTag()
	for _, tagStr := range tagList {
		tagStr = strings.TrimSpace(tagStr)
		if tagStr == "" {
			continue
		}

		v, ok := parseVersion(tagStr)
		if !ok {
			continue
		}

		versions = append(versions, v)
	}

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

	headHash, err := git.HeadHash()
	if err != nil {
		return err
	}
	latestHash, err := git.TagHash(latestVersion.String())
	if err != nil {
		return err
	}
	if headHash == latestHash {
		fmt.Println("already latest version")
		return nil
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

	err = git.CreatTag(newVersion.String())
	fmt.Println("tag created: " + newVersion.String())

	err = git.PushTag(newVersion.String())
	if err != nil {
		return err
	}
	fmt.Println("tag pushed")

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
