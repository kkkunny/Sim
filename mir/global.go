package mir

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/linkedlist"
	stlos "github.com/kkkunny/stl/os"
	"github.com/samber/lo"
)

type Global interface {
	Name()string
	Define()string
}

// 带名字结构体
type namedStruct struct {
	ctx Context
	index uint
	name string
	elems []Type
}

func (self *Module) NewNamedStructType(name string, elem ...Type)StructType {
	if _, ok := self.structMap[name]; ok{
		panic("unreachable")
	}
	for _, e := range elem{
		if !e.Context().Target().Equal(self.ctx.Target()){
			panic("unreachable")
		}
	}
	t := &namedStruct{
		ctx: self.ctx,
		name: name,
		elems: elem,
	}
	if name != ""{
		self.structMap[name] = t
	}
	self.globals.PushBack(t)
	return t
}

func (self *Module) NamedStructType(name string)(StructType, bool){
	res, ok := self.structMap[name]
	return res, ok
}

func (self *namedStruct) Name()string{
	if self.name != ""{
		return "t_" + self.name
	}
	return fmt.Sprintf("t_%d", self.index)
}

func (self *namedStruct) String()string{
	return self.Name()
}

func (self *namedStruct) Context()Context{
	return self.ctx
}

func (self *namedStruct) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*namedStruct)
	if !ok{
		return false
	}
	return self == dst
}

func (self *namedStruct) Align()stlos.Size{
	// TODO: get align
	return 1
}

func (self *namedStruct) Size()stlos.Size{
	// TODO: get size
	return 0
}

func (self *namedStruct) Elems()[]Type{
	return self.elems
}

func (self *namedStruct) Define()string{
	return fmt.Sprintf("type %s = %s", self.Name(), self)
}

// GlobalVariable 全局变量
type GlobalVariable struct {
	ctx Context
	index uint
	name string

	t Type
	mut bool
	value Const
}

func (self *Module) NewGlobalVariable(name string, t Type, value Const)*GlobalVariable{
	if _, ok := self.valueMap[name]; ok{
		panic("unreachable")
	}
	v := &GlobalVariable{
		ctx: self.ctx,
		name: name,
		t: t,
		value: value,
	}
	if name != ""{
		self.valueMap[name] = v
	}
	self.globals.PushBack(v)
	return v
}

func (self *GlobalVariable) Name()string{
	if self.name != ""{
		return "v_" + self.name
	}
	return fmt.Sprintf("v_%d", self.index)
}

func (self *GlobalVariable) Define()string{
	prefix := stlbasic.Ternary(self.mut, "var", "const")
	if self.value != nil{
		return fmt.Sprintf("%s %s %s = %s", prefix, self.t, self.Name(), self.value)
	}else{
		return fmt.Sprintf("%s %s %s", prefix, self.t, self.Name())
	}
}

func (self *GlobalVariable) Value()(Const, bool){
	return self.value, self.value!=nil
}

func (self *GlobalVariable) Type()Type{
	return self.t
}

func (self *GlobalVariable) String()string{
	return self.Name()
}

// Function 函数
type Function struct {
	ctx Context
	index uint
	name string

	t FuncType
	blocks linkedlist.LinkedList[Block]
}

func (self *Module) NewFunction(name string, t FuncType)*Function{
	if _, ok := self.valueMap[name]; ok{
		panic("unreachable")
	}
	v := &Function{
		ctx: self.ctx,
		name: name,
		t: t,
	}
	if name != ""{
		self.valueMap[name] = v
	}
	self.globals.PushBack(v)
	return v
}

func (self *Function) Name()string{
	if self.name != ""{
		return "f_" + self.name
	}
	return fmt.Sprintf("f_%d", self.index)
}

func (self *Function) Define()string{
	if self.blocks.Empty(){
		params := lo.Map(self.t.Params(), func(item Type, _ int) string {
			return item.String()
		})
		return fmt.Sprintf("%s %s(%s)", self.t.Ret(), self.Name(), strings.Join(params, ","))
	}

	var buf strings.Builder
	buf.WriteString(self.t.Ret().String())
	buf.WriteByte(' ')
	buf.WriteString(self.Name())
	buf.WriteByte('(')
	// TODO: params
	buf.WriteByte(')')
	return buf.String()
}

func (self *Function) Type()Type{
	return self.t
}

func (self *Function) String()string{
	return self.Name()
}
