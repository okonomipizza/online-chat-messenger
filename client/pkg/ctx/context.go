package context_setting

import (
	"context"
	"fmt"
	"net"
)

type ConnKeyType struct{}

func GetConnFromContext(ctx context.Context) (net.Conn, error) {
	conn, ok := ctx.Value(ConnKeyType{}).(net.Conn)
	if !ok {
		return nil, fmt.Errorf("no connection found in context")
	}
	return conn, nil
}
