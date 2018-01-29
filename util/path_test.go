package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitPath(t *testing.T) {
	dirs := SplitPath("/")
	assert.Equal(t, 2, len(dirs))
	assert.Equal(t, "", dirs[0])
	assert.Equal(t, "", dirs[1])

	dirs = SplitPath("/foo")
	assert.Equal(t, 2, len(dirs))
	assert.Equal(t, "", dirs[0])
	assert.Equal(t, "foo", dirs[1])

	dirs = SplitPath("/foo/bar")
	assert.Equal(t, 3, len(dirs))
	assert.Equal(t, "", dirs[0])
	assert.Equal(t, "foo", dirs[1])
	assert.Equal(t, "bar", dirs[2])

	dirs = SplitPath("/foo/bar/")
	assert.Equal(t, 4, len(dirs))
	assert.Equal(t, "", dirs[0])
	assert.Equal(t, "foo", dirs[1])
	assert.Equal(t, "bar", dirs[2])
	assert.Equal(t, "", dirs[3])
}
