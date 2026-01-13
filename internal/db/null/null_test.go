package null

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapString(t *testing.T) {
	s := WrapString("test")
	assert.Equal(t, "test", s.V)
	assert.Equal(t, true, s.Valid)

	s = WrapString("")
	assert.Equal(t, "", s.V)
	assert.Equal(t, false, s.Valid)

	i := WrapInt(10)
	assert.Equal(t, 10, i.V)
	assert.Equal(t, true, i.Valid)
	
	i = WrapInt(0)
	assert.Equal(t, 0, i.V)
	assert.Equal(t, false, i.Valid)
}
