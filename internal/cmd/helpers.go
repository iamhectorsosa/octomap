package cmd

import (
	"fmt"
)

const (
	MAX_NUM_ARGS   = 1
	invalidNumArgs = "accepts at most %d arg(s), received %d\n"
)

func validateRootArgs(args []string) error {
	if len(args) < MAX_NUM_ARGS {
		return nil
	}

	if len(args) > MAX_NUM_ARGS {
		return fmt.Errorf(invalidNumArgs, MAX_NUM_ARGS, len(args))
	}

	return nil
}
