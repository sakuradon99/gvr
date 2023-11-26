package git

import (
	"os/exec"
	"strings"
)

func gitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func HasChangeNotCommit() (bool, error) {
	raw, err := gitCommand("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return raw != "", nil
}

func HasCommitNotPushed() (bool, error) {
	raw, err := gitCommand("cherry", "-v")
	if err != nil {
		return false, err
	}
	return raw != "", nil
}

func FetchTags() error {
	_, err := gitCommand("fetch", "--tags")
	return err
}

func ListTag() ([]string, error) {
	raw, err := gitCommand("tag")
	if err != nil {
		return nil, err
	}

	return strings.Split(raw, "\n"), nil
}

func CreatTag(name string) error {
	_, err := gitCommand("tag", name)
	return err
}

func TagHash(name string) (string, error) {
	return gitCommand("rev-list", "-n", "1", name)
}

func HeadHash() (string, error) {
	return gitCommand("rev-parse", "HEAD")
}

func PushTag(name string) error {
	_, err := gitCommand("push", "origin", name)
	if err != nil {
		return err
	}
	return nil
}
