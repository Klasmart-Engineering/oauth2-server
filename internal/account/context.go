package account

import "context"

type ctxKey struct{}

var accountCtxKey = ctxKey{}

func GetAccountIdFromCtx(ctx context.Context) string {
	return ctx.Value(accountCtxKey).(string)
}
