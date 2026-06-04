package core

import "context"

// sessionKeyCtxKeyT is an unexported context key type so values set here can't
// collide with context values from other packages.
type sessionKeyCtxKeyT struct{}

var sessionKeyCtxKey = sessionKeyCtxKeyT{}

// WithSessionKey returns a child context carrying the cc-connect session key
// (e.g. "feishu:{chatID}:{threadRootID}"). It is set per StartSession call so
// agents can route per-session resources (e.g. a git worktree per chat topic)
// without mutating shared agent state — context values are immutable and
// per-call, so concurrent topics never race.
func WithSessionKey(ctx context.Context, key string) context.Context {
	if key == "" {
		return ctx
	}
	return context.WithValue(ctx, sessionKeyCtxKey, key)
}

// SessionKeyFromContext returns the session key set by WithSessionKey, or "".
func SessionKeyFromContext(ctx context.Context) string {
	v, _ := ctx.Value(sessionKeyCtxKey).(string)
	return v
}
