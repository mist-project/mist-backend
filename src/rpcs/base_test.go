package rpcs_test

import (
	"fmt"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	"github.com/stretchr/testify/assert"

	"mist/src/rpcs"
	"mist/src/testutil"
)

func TestBaseInterceptors_Success(t *testing.T) {
	t.Run("Successcreating_interceptors_does_not_fail", func(t *testing.T) {
		opt, err := rpcs.BaseInterceptors()
		assert.NotNil(t, opt)
		assert.Nil(t, err)
	})

	t.Run("Successcreating_interceptors_does_not_fail", func(t *testing.T) {
		// ARRANGE
		mockValidator := new(testutil.MockProtovalidator)
		// Backup and override the global validator creator
		original := rpcs.NewValidator

		rpcs.NewValidator = func() (protovalidate.Validator, error) {
			return mockValidator, fmt.Errorf("fail creating validator")
		}

		t.Cleanup(func() {
			rpcs.NewValidator = original
		})

		// ACT
		_, err := rpcs.BaseInterceptors()

		// ASSERT
		assert.NotNil(t, err)
	})

}
