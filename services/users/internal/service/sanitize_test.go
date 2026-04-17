package service

import "testing"

func TestEscapeLikePattern(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello",        "hello"},
		{"%",            `\%`},
		{"_",            `\_`},
		{`\`,            `\\`},
		{"%_\\",         `\%\_\\`},
		{"foo%bar",      `foo\%bar`},
		{"foo_bar",      `foo\_bar`},
		{"normal input", "normal input"},
		// Classic injection attempts should be stored literally
		{"' OR '1'='1", "' OR '1'='1"},
		{"%' OR 1=1 --", `\%' OR 1=1 --`},
	}

	for _, tc := range tests {
		got := EscapeLikePattern(tc.input)
		if got != tc.want {
			t.Errorf("EscapeLikePattern(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}
