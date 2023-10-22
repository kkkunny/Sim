package parse

import "github.com/kkkunny/Sim/token"

func loopParseWithUtil[T any](self *Parser, sem, end token.Kind, f func() T) (res []T) {
	for self.skipSEM(); !self.nextIs(end); self.skipSEM() {
		res = append(res, f())
		if !self.skipNextIs(sem) {
			break
		}
	}
	return res
}
