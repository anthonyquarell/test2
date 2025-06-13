package comportal

import "context"

type Client interface {
	Send(ctx context.Context, obj *SendReq) (_ []byte, finalError error)
}
