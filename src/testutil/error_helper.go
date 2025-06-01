package testutil

import (
	"mist/src/faults"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertCustomErrorContains(t *testing.T, err error, expected string) {
	customErr, ok := err.(*faults.CustomError)

	if !ok {
		assert.Equal(t, err.Error(), expected)
	} else {
		assert.Contains(t, customErr.StackTrace(), expected)
	}
}
