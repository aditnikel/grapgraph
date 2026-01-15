package redis

import (
	"context"

	"github.com/redis/rueidis"
)

type Client struct {
	RDB    rueidis.Client
	Ctx    context.Context
	Ledger string
}

func New(addr, ledger string) (*Client, error) {
	rdb, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{addr},
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		RDB:    rdb,
		Ctx:    context.Background(),
		Ledger: ledger,
	}, nil
}
