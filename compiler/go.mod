module github.com/kkkunny/Sim

go 1.21.3

replace github.com/kkkunny/Sim/runtime v0.0.0-20231203140042-fdf7df5c03c7 => ../runtime

require (
	github.com/kkkunny/Sim/runtime v0.0.0-20231203140042-fdf7df5c03c7
	github.com/kkkunny/stl v0.0.0-20231114084730-4ad25211cae2
)

require (
	github.com/kkkunny/go-llvm v0.0.0-20231206063638-5a0b53f41dfd
	github.com/samber/lo v1.38.1
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
)

require (
	github.com/gookit/color v1.5.4 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/sys v0.12.0 // indirect
)
