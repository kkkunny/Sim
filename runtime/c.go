package runtime

import "C"
import "os"

// Exit 退出程序
func Exit(code C.int) {
	os.Exit(int(code))
}
