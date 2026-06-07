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

	// Two distinct thread keys for the SAME chat differ only in the tail of the
	// root message id. The fixed prefix (feishu_ + 35-char oc_ chat id + _root_)
	// alone eats ~48 chars, so a naive truncate-at-60 throws away the only
	// distinguishing bytes and collides both threads onto one worktree.
	// Regression for: two Feishu topics routed to the same worktree.
	chat := "feishu:oc_c3e20d7dba637a4cf3c621c2b334f96d:root:om_x100b6d71"
	keyA := chat + "aaaaaaaaaaaa"
	keyB := chat + "bbbbbbbbbbbb"
	slugA, slugB := sessionSlug(keyA), sessionSlug(keyB)
	if slugA == slugB {
		t.Errorf("distinct thread keys collided onto one slug: %q", slugA)
	}
	if len(slugA) > 60 || len(slugB) > 60 {
		t.Errorf("slug not truncated: lenA=%d lenB=%d", len(slugA), len(slugB))
	}
	// Slug must be a pure function of the key (reuse depends on it being stable
	// across messages/process restarts).
	if sessionSlug(keyA) != slugA {
		t.Errorf("sessionSlug not deterministic for %q", keyA)
	}
}

func TestLastNonEmptyLine(t *testing.T) {
	cases := map[string]string{
		"/path/to/wt\n":           "/path/to/wt",
		"log line\n/path/to/wt\n": "/path/to/wt",
		"  /path/with/spaces  \n": "/path/with/spaces",
		"":                        "",
		"\n\n":                    "",
	}
	for in, want := range cases {
		if got := lastNonEmptyLine(in); got != want {
			t.Errorf("lastNonEmptyLine(%q) = %q, want %q", in, got, want)
		}
	}
}
