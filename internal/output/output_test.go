package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rodmhgl/ldctl/internal/output"
)

// ---------------------------------------------------------------------------
// PrintJSON
// ---------------------------------------------------------------------------

func TestPrintJSON_SimpleStruct(t *testing.T) {
	t.Parallel()

	type item struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	var buf bytes.Buffer
	err := output.PrintJSON(&buf, item{Name: "go", Count: 42})
	require.NoError(t, err)

	got := buf.String()
	assert.Contains(t, got, `"name": "go"`)
	assert.Contains(t, got, `"count": 42`)
	// must end with a newline
	assert.True(t, strings.HasSuffix(got, "\n"), "expected trailing newline")
}

func TestPrintJSON_TwoSpaceIndent(t *testing.T) {
	t.Parallel()

	type inner struct {
		X int `json:"x"`
	}

	var buf bytes.Buffer
	err := output.PrintJSON(&buf, inner{X: 1})
	require.NoError(t, err)

	// 2-space indent means lines like "  \"x\": 1"
	assert.Contains(t, buf.String(), "  \"x\": 1")
}

func TestPrintJSON_NilValueRendersNull(t *testing.T) {
	t.Parallel()

	type maybeStr struct {
		Val *string `json:"val"`
	}

	var buf bytes.Buffer
	err := output.PrintJSON(&buf, maybeStr{Val: nil})
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `"val": null`)
}

func TestPrintJSON_EmptySliceRendersEmptyArray(t *testing.T) {
	t.Parallel()

	type withSlice struct {
		Tags []string `json:"tags"`
	}

	var buf bytes.Buffer
	err := output.PrintJSON(&buf, withSlice{Tags: []string{}})
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `"tags": []`)
}

func TestPrintJSON_ValidJSON(t *testing.T) {
	t.Parallel()

	type payload struct {
		A string `json:"a"`
		B int    `json:"b"`
	}

	var buf bytes.Buffer
	err := output.PrintJSON(&buf, payload{A: "hello", B: 7})
	require.NoError(t, err)

	// Verify the output is actually parseable JSON.
	var decoded payload
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, "hello", decoded.A)
	assert.Equal(t, 7, decoded.B)
}

func TestPrintJSON_Map(t *testing.T) {
	t.Parallel()

	data := map[string]any{
		"key": "value",
	}

	var buf bytes.Buffer
	err := output.PrintJSON(&buf, data)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `"key": "value"`)
}

func TestPrintJSON_Slice(t *testing.T) {
	t.Parallel()

	data := []string{"alpha", "beta"}

	var buf bytes.Buffer
	err := output.PrintJSON(&buf, data)
	require.NoError(t, err)

	got := buf.String()
	assert.Contains(t, got, `"alpha"`)
	assert.Contains(t, got, `"beta"`)
}

func TestPrintJSON_ErrorOnUnmarshalable(t *testing.T) {
	t.Parallel()

	// Channels cannot be marshalled to JSON.
	ch := make(chan int)
	var buf bytes.Buffer
	err := output.PrintJSON(&buf, ch)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// PrintData
// ---------------------------------------------------------------------------

func TestPrintData_WritesFormattedString(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintData(&buf, "hello %s, count=%d\n", "world", 3)

	assert.Equal(t, "hello world, count=3\n", buf.String())
}

func TestPrintData_NoArgs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintData(&buf, "plain text")

	assert.Equal(t, "plain text", buf.String())
}

// ---------------------------------------------------------------------------
// PrintError
// ---------------------------------------------------------------------------

func TestPrintError_PrefixesWithError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintError(&buf, "something went wrong: %s\n", "oops")

	assert.Equal(t, "Error: something went wrong: oops\n", buf.String())
}

func TestPrintError_NoArgs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintError(&buf, "fatal\n")

	assert.Equal(t, "Error: fatal\n", buf.String())
}

// ---------------------------------------------------------------------------
// PrintProgress
// ---------------------------------------------------------------------------

func TestPrintProgress_WhenNotQuiet_WritesOutput(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintProgress(&buf, false, "Processing %d items\n", 5)

	assert.Equal(t, "Processing 5 items\n", buf.String())
}

func TestPrintProgress_WhenQuiet_WritesNothing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintProgress(&buf, true, "you should not see this\n")

	assert.Empty(t, buf.String())
}

func TestPrintProgress_MultipleCallsQuietMixed(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintProgress(&buf, true, "suppressed\n")
	output.PrintProgress(&buf, false, "visible\n")
	output.PrintProgress(&buf, true, "suppressed again\n")

	assert.Equal(t, "visible\n", buf.String())
}

// ---------------------------------------------------------------------------
// PrintVerbose
// ---------------------------------------------------------------------------

func TestPrintVerbose_WhenVerbose_WritesDebugPrefix(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintVerbose(&buf, true, "connecting to %s\n", "localhost")

	assert.Equal(t, "[DEBUG] connecting to localhost\n", buf.String())
}

func TestPrintVerbose_WhenNotVerbose_WritesNothing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintVerbose(&buf, false, "this should be hidden\n")

	assert.Empty(t, buf.String())
}

func TestPrintVerbose_AlwaysPrefixesWithDebug(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintVerbose(&buf, true, "step 1\n")
	output.PrintVerbose(&buf, true, "step 2\n")

	got := buf.String()
	assert.Equal(t, "[DEBUG] step 1\n[DEBUG] step 2\n", got)
}

func TestPrintVerbose_NoArgs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	output.PrintVerbose(&buf, true, "no format args\n")

	assert.Equal(t, "[DEBUG] no format args\n", buf.String())
}
