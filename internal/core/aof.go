package core

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/resp"
)

func dumpKey(fp *os.File, key string, obj *Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	fp.Write(resp.Encode(tokens, false))
}

func DumpAllAOF() {
	fp, err := os.OpenFile(config.AOFFile, os.O_CREATE | os.O_WRONLY, os.ModeAppend)

	if err != nil {
		fmt.Print("error", err)
		return
	}

	log.Println("Rewriting AOF file at ", config.AOFFile)
	for k, obj := range store {
		dumpKey(fp, k, obj)
	}
	log.Println("AOF file rewrite complete.")
}