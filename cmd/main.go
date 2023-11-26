package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"regexp"
	"sort"
	"strconv"
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
	repo, err := git.PlainOpen("./")
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	status, err := worktree.Status()
	if err != nil {
		return err
	}

	if !status.IsClean() {
		fmt.Println("has uncommitted changes")
		return nil
	}

	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Tags:       git.AllTags,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	tags, err := repo.Tags()
	if err != nil {
		return err
	}

	var versions []version
	tagToRef := make(map[string]*plumbing.Reference)
	_ = tags.ForEach(func(ref *plumbing.Reference) error {
		tag := ref.Name().Short()
		v, ok := parseVersion(tag)
		if !ok {
			return nil
		}
		versions = append(versions, v)
		tagToRef[v.String()] = ref
		return nil
	})

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

	currentRef, err := repo.Head()
	if err != nil {
		return err
	}
	latestVersionRef := tagToRef[latestVersion.String()]
	if latestVersionRef != nil && latestVersionRef.Hash() == currentRef.Hash() {
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

	tagRef, err := repo.CreateTag(newVersion.String(), currentRef.Hash(), &git.CreateTagOptions{})
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("tag <%s> created, hash: %s", tagRef.Name().Short(), tagRef.Hash()))

	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tagRef.Name().Short(), tagRef.Name().Short())),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func handleError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, git.ErrRepositoryNotExists) {
		fmt.Println("not a git repository")
		return nil
	}
	if errors.Is(err, git.ErrRemoteNotFound) {
		fmt.Println("remote not found")
	}

	return err
}

var versionRegexp = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

type version struct {
	major int
	minor int
	patch int
	ref   *plumbing.Reference
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