package cmd

import (
	"fmt"
)

const (
	maxNumArgs     = 1
	invalidNumArgs = "accepts at most %d arg(s), received %d\n"
)

func validateRootArgs(args []string) error {
	if len(args) < maxNumArgs {
		return nil
	}

	if len(args) > maxNumArgs {
		return fmt.Errorf(invalidNumArgs, maxNumArgs, len(args))
	}

	return nil
}
