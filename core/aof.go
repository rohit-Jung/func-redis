package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/rohit-Jung/func-redis/config"
)

func dumpKey(fp *os.File, k string, obj *Obj) {
	cmd := fmt.Sprintf("SET %s %s", k, obj.Value)
	tokens := strings.Split(cmd, " ")
	fmt.Println("Tokens", tokens)
	fp.Write(Encode(tokens, false))
}

func DumpAllAof() {
	fp, err := os.OpenFile(config.AofFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	for k, v := range store {
		dumpKey(fp, k, v)
	}
}
