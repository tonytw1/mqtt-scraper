package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCanParseIntegerNumbers(t *testing.T) {
	value, err := parseMaybeNumber("123")
	require.Nil(t, err)

	assert.Equal(t, "123", value.String())
}

func TestCanParseFloatNumbers(t *testing.T) {
	value, err := parseMaybeNumber("123.46")
	require.Nil(t, err)

	assert.Equal(t, "123.46", value.String())
}
