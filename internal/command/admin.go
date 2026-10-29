package command

import (
	"fmt"

	"github.com/mnnbnsl/sider/internal/persistence"
	"github.com/mnnbnsl/sider/internal/resp"
)

func EvalBGREWRITEAOF(args []string) []byte {
	if err := persistence.RewriteAOF(); err != nil {
		return resp.Encode(fmt.Errorf("ERR failed to rewrite AOF: %w", err), false)
	}
	return resp.RESP_OK
}
