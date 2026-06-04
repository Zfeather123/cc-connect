package claudecode

import "testing"

func TestSessionSlug(t *testing.T) {
	cases := map[string]string{
		"feishu:oc_abc:om_123": "feishu_oc_abc_om_123",
		"FEISHU:OC_ABC":        "feishu_oc_abc",
		"::feishu::":           "feishu",
		"a/b\\c.d":             "a_b_c_d",
		"":                     "",
	}
	for in, want := range cases {
		if got := sessionSlug(in); got != want {
			t.Errorf("sessionSlug(%q) = %q, want %q", in, got, want)
		}
	}

	// Long keys are truncated to <=60 and not left with a trailing underscore.
	long := ""
	for i := 0; i < 100; i++ {
		long += "x"
	}
	if got := sessionSlug(long + ":" + long); len(got) > 60 {
		t.Errorf("slug not truncated: len=%d", len(got))
	}
}

func TestLastNonEmptyLine(t *testing.T) {
	cases := map[string]string{
		"/path/to/wt\n":            "/path/to/wt",
		"log line\n/path/to/wt\n":  "/path/to/wt",
		"  /path/with/spaces  \n":  "/path/with/spaces",
		"":                         "",
		"\n\n":                     "",
	}
	for in, want := range cases {
		if got := lastNonEmptyLine(in); got != want {
			t.Errorf("lastNonEmptyLine(%q) = %q, want %q", in, got, want)
		}
	}
}
