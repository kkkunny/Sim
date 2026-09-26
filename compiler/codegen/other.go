package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/globals"
)

// stableName 获取固定的符号名。
// 用相对 SimRootPath 的包路径参与哈希（而非绝对路径），使编译缓存可随工作区迁移；
// 位于 SimRootPath 之外的包（如直接分析 /tmp 下的单文件）回退为绝对路径。
func stableName(pkg *globals.Package, name string) string {
	pkgPath := pkg.Path
	if rel, err := filepath.Rel(config.SimRootPath, pkg.Path); err == nil && !strings.HasPrefix(rel, "..") {
		pkgPath = filepath.ToSlash(rel)
	}

	h := sha256.New()
	h.Write([]byte(config.ABIVersion))
	h.Write([]byte{0})
	h.Write([]byte(pkgPath))
	h.Write([]byte{0})
	h.Write([]byte(pkg.Name))
	digest := h.Sum(nil)
	return fmt.Sprintf("_Sim_%s_%s", hex.EncodeToString(digest)[:16], name)
}
