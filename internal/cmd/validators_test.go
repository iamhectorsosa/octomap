package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRootArgs(t *testing.T) {
	tests := []struct {
		name string
		err  error
		args []string
	}{
		{
			name: "No arguments",
			args: []string{},
		},
		{
			name: "One arguent",
			args: []string{"arg1"},
		},
		{
			name: "More than one argument",
			args: []string{"arg1", "arg2"},
			err:  fmt.Errorf(invalidNumArgs, maxNumArgs, 2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRootArgs(tt.args)
			if tt.err != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
