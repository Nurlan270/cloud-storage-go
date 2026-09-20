package context

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
)

var txCtxKey = &contextKey{"tx"}

func NewTxContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey, tx)
}

func TxFromContext(ctx context.Context) pgx.Tx {
	return ctx.Value(txCtxKey).(pgx.Tx)
}

func TxFromRequest(r *http.Request) pgx.Tx {
	return TxFromContext(r.Context())
}
