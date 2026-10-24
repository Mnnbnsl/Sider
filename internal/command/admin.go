package command

import (
	"github.com/mnnbnsl/sider/internal/persistence"
	"github.com/mnnbnsl/sider/internal/resp"
)

func EvalBGREWRITEAOF(args []string) []byte {
	persistence.DumpAllAOF()
	return resp.RESP_OK
}
