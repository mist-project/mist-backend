package testutil

import (
	"github.com/bufbuild/protovalidate-go"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

type MockProtovalidator struct {
	mock.Mock
}

func (m *MockProtovalidator) Validate(msg proto.Message, options ...protovalidate.ValidationOption) error {
	args := m.Called(msg)
	return args.Error(0)
}
