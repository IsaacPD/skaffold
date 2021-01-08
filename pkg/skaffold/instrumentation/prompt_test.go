package instrumentation

import (
	"bytes"
	"io"
	"testing"

	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestDisplaySurveyForm(t *testing.T) {
	tests := []struct {
		description string
		mockStdOut  bool
		expected    string
	}{
		{
			description: "std out",
			mockStdOut:  true,
			expected:    Prompt,
		},
		{
			description: "not std out",
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			mock := func(io.Writer) bool { return test.mockStdOut }
			t.Override(&isStdOut, mock)
			t.Override(&updateConfig, func(_ string) error { return nil })
			var buf bytes.Buffer
			err := DisplayMetricsPrompt("", &buf)
			t.CheckErrorAndDeepEqual(false, err, test.expected, buf.String())
		})
	}
}
