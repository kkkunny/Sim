package analyse

const (
	errDuplicateDeclaration    = "duplicate declaration"                           // 重复声明
	errUnknownIdentifier       = "unknown identifier"                              // 未知的标识符
	errDataOverflow            = "data overflow"                                   // 数据溢出
	errCircularReference       = "circular reference"                              // 循环引用
	errNotExpectImmediateValue = "not expect a immediate value"                    // 不希望是一个立即数
	errExpectMutableValue      = "expect a mutable value"                          // 希望是一个可变值
	errExpectBoolean           = "expect a boolean"                                // 期待一个布尔值
	errExpectNumber            = "expect a number"                                 // 期待一个数字
	errExpectInteger           = "expect a integer"                                // 期待一个整数
	errExpectUnsafeBlock       = "expect in a unsafe block"                        // 期待处于一个不安全的作用域中
	errNotSatisfyConstraint    = "failure to satisfy constraint for generic param" // 不能满足泛型约束
)
