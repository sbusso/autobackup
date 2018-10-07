package sources

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseArg(t *testing.T) {
	r := require.New(t)

	long := []string{"--foo", "--bar", "1234", "--baz"}
	short := []string{"--foo", "-b=1234", "--baz"}

	res := parseArg(long, "--bar")
	r.NotContains(res, "1234")

	res = parseArg(long, "--none")
	r.Equal(res, long)

	res = parseArg(short, "-b")
	r.NotContains(res, "-b=1234")

	res = parseArg(short, "-c")
	r.Contains(res, "-b=1234")

	res = parseArg(long, "")
	r.Equal(res, long)
}
