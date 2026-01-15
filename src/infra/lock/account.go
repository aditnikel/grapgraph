package lock

import (
	"context"

	rueidis "github.com/redis/rueidis"
)

func Lock(r rueidis.Client, ctx context.Context, id string) bool {
	return r.Do(ctx,
		r.B().Set().
			Key("lock:acct:"+id).
			Value("1").
			Nx().
			Ex(5).
			Build(),
	).Error() == nil
}
