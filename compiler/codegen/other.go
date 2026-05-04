package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

// 获取固定的name
func stableName(pkg *stmts.Package, name string) string {
	h := sha256.New()
	h.Write([]byte(config.ABIVersion))
	h.Write([]byte{0})
	h.Write([]byte(pkg.Path))
	h.Write([]byte{0})
	h.Write([]byte(pkg.Name))
	digest := h.Sum(nil)
	return fmt.Sprintf("_Sim_%s_%s", hex.EncodeToString(digest)[:16], name)
}
