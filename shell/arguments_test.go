package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArguments(t *testing.T) {
	assert.Equal(t, []string{"foo", "bar", "baz"}, parseArguments("foo bar baz"))
	assert.Equal(t, []string{"foo", "bar", "baz"}, parseArguments(" foo   bar baz  "))
	assert.Equal(t, []string{"foo", "bar\\ baz"}, parseArguments(" foo  bar\\ baz  "))
	assert.Equal(t, []string{"foo", "bar\\ baz", "bax"}, parseArguments(" foo  bar\\ baz bax"))

}
