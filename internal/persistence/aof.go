package persistence

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mnnbnsl/sider/config"
	"github.com/mnnbnsl/sider/internal/resp"
	"github.com/mnnbnsl/sider/internal/store"
)

func dumpKey(fp *os.File, key string, obj *store.Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	fp.Write(resp.Encode(tokens, false))
}

func DumpAllAOF() {
	fp, err := os.OpenFile(config.AOFFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("error", err)
		return
	}
	defer fp.Close()

	log.Println("Rewriting AOF file at ", config.AOFFile)
	store.Range(func(key string, obj *store.Obj) {
		dumpKey(fp, key, obj)
	})
	log.Println("AOF file rewrite complete.")
}
