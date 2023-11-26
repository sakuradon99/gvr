package git

import (
	"os/exec"
)

func Status() (string, error) {
	cmd := exec.Command("git", "status")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
