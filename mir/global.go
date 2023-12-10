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
	Context()*Context
	Name()string
	Define()string
	setIndex(i uint)
}

// 带名字结构体
type namedStruct struct {
	ctx *Context
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

func (self *namedStruct) setIndex(i uint){
	self.index = i
}

func (self *namedStruct) Name()string{
	if self.name != ""{
		return "t_" + self.name
	}
	return fmt.Sprintf("t%d", self.index)
}

func (self *namedStruct) String()string{
	return self.Name()
}

func (self *namedStruct) Context()*Context{
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

func (self *namedStruct) SetElems(elem ...Type){
	self.elems = elem
}

func (self *namedStruct) Elems()[]Type{
	return self.elems
}

func (self *namedStruct) Define()string{
	return fmt.Sprintf("type %s = %s", self.Name(), self)
}

// GlobalVariable 全局变量
type GlobalVariable struct {
	ctx *Context
	index uint
	name string

	t Type
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

func (self *Module) NamedGlobalVariable(name string)(*GlobalVariable, bool){
	v, ok := self.valueMap[name]
	if !ok || !stlbasic.Is[*GlobalVariable](v){
		return nil, false
	}
	return v.(*GlobalVariable), true
}

func (self *GlobalVariable) setIndex(i uint){
	self.index = i
}

func (self *GlobalVariable) Name()string{
	if self.name != ""{
		return "v_" + self.name
	}
	return fmt.Sprintf("v%d", self.index)
}

func (self *GlobalVariable) Context()*Context{
	return self.ctx
}

func (self *GlobalVariable) Define()string{
	if self.value != nil{
		return fmt.Sprintf("var %s %s = %s", self.t, self.Name(), self.value)
	}else{
		return fmt.Sprintf("var %s %s", self.t, self.Name())
	}
}

func (self *GlobalVariable) SetValue(v Const){
	self.value = v
}

func (self *GlobalVariable) Value()(Const, bool){
	return self.value, self.value!=nil
}

func (self *GlobalVariable) Type()Type{
	return self.ctx.NewPtrType(self.t)
}

func (self *GlobalVariable) ValueType()Type{
	return self.t
}

// Function 函数
type Function struct {
	ctx *Context
	index uint
	name string

	t FuncType
	params []*Param
	blocks linkedlist.LinkedList[*Block]
}

func (self *Module) NewFunction(name string, t FuncType)*Function{
	if _, ok := self.valueMap[name]; ok{
		panic("unreachable")
	}
	v := &Function{
		ctx: self.ctx,
		name: name,
		t: t,
		params: lo.Map(t.Params(), func(item Type, index int) *Param {
			return newParam(uint(index), item)
		}),
	}
	if name != ""{
		self.valueMap[name] = v
	}
	self.globals.PushBack(v)
	return v
}

func (self *Module) NamedFunction(name string)(*Function, bool){
	v, ok := self.valueMap[name]
	if !ok || !stlbasic.Is[*Function](v){
		return nil, false
	}
	return v.(*Function), true
}

func (self *Function) setIndex(i uint){
	self.index = i
}

func (self *Function) Name()string{
	if self.name != ""{
		return "f_" + self.name
	}
	return fmt.Sprintf("f%d", self.index)
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
	for i, p := range self.params{
		buf.WriteString(p.Define())
		if i < len(self.params) - 1{
			buf.WriteByte(',')
		}
	}
	buf.WriteString(")\n")
	var i uint
	for iter:=self.blocks.Iterator(); iter.Next(); {
		iter.Value().index = i
		buf.WriteString(iter.Value().String())
		if iter.HasNext(){
			buf.WriteByte('\n')
		}
		i++
	}
	return buf.String()
}

func (self *Function) Type()Type{
	return self.t
}

func (self *Function) Context()*Context{
	return self.ctx
}

func (self *Function) Params()[]*Param{
	return self.params
}

func (self *Function) Blocks()linkedlist.LinkedList[*Block]{
	return self.blocks
}

// Constant 常量
type Constant struct {
	ctx *Context
	index uint
	name string

	value Const
}

func (self *Module) NewConstant(name string, value Const)*Constant{
	if _, ok := self.valueMap[name]; ok{
		panic("unreachable")
	}
	if value == nil{
		panic("unreachable")
	}
	v := &Constant{
		ctx: self.ctx,
		name: name,
		value: value,
	}
	if name != ""{
		self.valueMap[name] = v
	}
	self.globals.PushBack(v)
	return v
}

func (self *Module) NamedConstant(name string)(*Constant, bool){
	v, ok := self.valueMap[name]
	if !ok || !stlbasic.Is[*Constant](v){
		return nil, false
	}
	return v.(*Constant), true
}

func (self *Constant) setIndex(i uint){
	self.index = i
}

func (self *Constant) Name()string{
	if self.name != ""{
		return "c_" + self.name
	}
	return fmt.Sprintf("c%d", self.index)
}

func (self *Constant) Context()*Context{
	return self.ctx
}

func (self *Constant) Define()string{
	return fmt.Sprintf("const %s %s = %s", self.Type(), self.Name(), self.value)
}

func (self *Constant) Value()Const{
	return self.value
}

func (self *Constant) Type()Type{
	return self.ctx.NewPtrType(self.ValueType())
}

func (self *Constant) IsZero()bool{
	return self.value.IsZero()
}

func (*Constant) constant(){}

func (self *Constant) ValueType()Type{
	return self.value.Type()
}
