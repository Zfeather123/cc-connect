package claudecode

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var worktreeSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

// sessionSlug turns a session key like "feishu:oc_xxx:om_yyy" into a
// filesystem- and git-branch-safe slug.
func sessionSlug(key string) string {
	s := strings.ToLower(key)
	s = worktreeSlugRe.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if len(s) > 60 {
		// A plain s[:60] would drop the distinguishing tail: the fixed prefix
		// (feishu_ + 35-char oc_ chat id + _root_) alone eats ~48 chars, so two
		// thread keys in the same chat survive only by the head of their root
		// message id and collide. Keep a readable prefix and append a hash of the
		// FULL key to restore uniqueness while staying deterministic.
		sum := sha1.Sum([]byte(key))
		s = strings.Trim(s[:51], "_") + "_" + hex.EncodeToString(sum[:])[:8]
	}
	return s
}

// ensureWorktree returns the working directory for a given session key,
// creating an isolated git worktree (with its own DB/ports) on first use and
// reusing it afterward. Provisioning is delegated to
// <base>/.claude/hooks/cc-connect-worktree.sh when present (so DB/port logic
// lives in an editable script, not compiled Go); otherwise it falls back to a
// plain `git worktree add`. On any error it returns base (the shared work_dir)
// so the session still runs rather than failing outright.
func (a *Agent) ensureWorktree(base, key string) (string, error) {
	slug := sessionSlug(key)
	if slug == "" {
		return base, fmt.Errorf("empty slug for session key %q", key)
	}

	// Reuse a previously-resolved worktree if it still exists on disk.
	if v, ok := a.worktreeCache.Load(slug); ok {
		if p, _ := v.(string); p != "" {
			if st, err := os.Stat(p); err == nil && st.IsDir() {
				return p, nil
			}
			a.worktreeCache.Delete(slug)
		}
	}

	script := filepath.Join(base, ".claude", "hooks", "cc-connect-worktree.sh")
	var path string
	if st, err := os.Stat(script); err == nil && !st.IsDir() {
		payload, _ := json.Marshal(map[string]string{"name": slug, "session_key": key})
		cmd := exec.Command("bash", script)
		cmd.Dir = base
		cmd.Stdin = bytes.NewReader(payload)
		var out, errBuf bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errBuf
		if err := cmd.Run(); err != nil {
			return base, fmt.Errorf("cc-connect-worktree.sh failed: %w; stderr: %s", err, strings.TrimSpace(errBuf.String()))
		}
		path = lastNonEmptyLine(out.String())
		if path == "" {
			return base, fmt.Errorf("cc-connect-worktree.sh produced no path; stderr: %s", strings.TrimSpace(errBuf.String()))
		}
	} else {
		// Fallback: plain git worktree, no DB/port isolation.
		path = filepath.Join(base, ".claude", "worktrees", slug)
		if st, err := os.Stat(path); err != nil || !st.IsDir() {
			cmd := exec.Command("git", "worktree", "add", path, "-b", "worktree-"+slug)
			cmd.Dir = base
			if out, err := cmd.CombinedOutput(); err != nil {
				return base, fmt.Errorf("git worktree add failed: %w; %s", err, strings.TrimSpace(string(out)))
			}
		}
	}

	a.worktreeCache.Store(slug, path)
	slog.Info("claudecode: routed session to worktree", "session_key", key, "slug", slug, "work_dir", path)
	return path, nil
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if t := strings.TrimSpace(lines[i]); t != "" {
			return t
		}
	}
	return ""
}
