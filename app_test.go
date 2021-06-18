package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCanParseIntegerNumbers(t *testing.T) {
	value, err := parseMaybeNumber("123")
	require.Nil(t, err)

	f, _ := value.Float64()
	assert.Equal(t, 123.0, f)
}

func TestCanParseFloatNumbers(t *testing.T) {
	value, err := parseMaybeNumber("123.46")
	require.Nil(t, err)

	f, _ := value.Float64()
	assert.Equal(t, 123.46, f)
}

func TestCanSmallFloats(t *testing.T) {
	value, err := parseMaybeNumber("0.123")
	require.Nil(t, err)

	f, _ := value.Float64()
	assert.Equal(t, 0.123, f)
}
