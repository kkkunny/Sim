package mir

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/list"
	stlos "github.com/kkkunny/stl/os"
	"github.com/samber/lo"
)

type Global interface {
	Context()*Context
	Name()string
	RealName()string
	Define()string
	setIndex(i uint) uint
}

// NamedStruct 带名字结构体
type NamedStruct struct {
	ctx *Context
	index uint
	name string
	elems []Type
}

func (self *Module) NewNamedStructType(name string, elem ...Type)*NamedStruct {
	if _, ok := self.structMap[name]; ok{
		panic("unreachable")
	}
	for _, e := range elem{
		if !e.Context().Target().Equal(self.ctx.Target()){
			panic("unreachable")
		}
	}
	t := &NamedStruct{
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

func (self *Module) NamedStructType(name string)(*NamedStruct, bool){
	res, ok := self.structMap[name]
	return res, ok
}

func (self *NamedStruct) setIndex(i uint) uint {
	if self.name != ""{
		return i
	}
	self.index = i
	return i+1
}

func (self *NamedStruct) Name()string{
	if self.name != ""{
		return "t_" + self.name
	}
	return fmt.Sprintf("t%d", self.index)
}

func (self *NamedStruct) RealName()string{
	return self.name
}

func (self *NamedStruct) String()string{
	return self.Name()
}

func (self *NamedStruct) Context()*Context{
	return self.ctx
}

func (self *NamedStruct) Equal(t Type)bool{
	if !self.ctx.Target().Equal(t.Context().Target()){
		return false
	}
	dst, ok := t.(*NamedStruct)
	if !ok{
		return false
	}
	return self == dst
}

func (self *NamedStruct) Align() uint64 {
	return self.ctx.target.AlignOf(self)
}

func (self *NamedStruct) Size()stlos.Size{
	return self.ctx.target.SizeOf(self)
}

func (self *NamedStruct) SetElems(elem ...Type){
	self.elems = elem
}

func (self *NamedStruct) Elems()[]Type{
	return self.elems
}

func (self *NamedStruct) Define()string{
	elems := lo.Map(self.elems, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("type %s = {%s}", self.Name(), strings.Join(elems, ","))
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

func (self *GlobalVariable) setIndex(i uint) uint {
	if self.name != ""{
		return i
	}
	self.index = i
	return i+1
}

func (self *GlobalVariable) Name()string{
	if self.name != ""{
		return "v_" + self.name
	}
	return fmt.Sprintf("v%d", self.index)
}

func (self *GlobalVariable) RealName()string{
	return self.name
}

func (self *GlobalVariable) Context()*Context{
	return self.ctx
}

func (self *GlobalVariable) Define()string{
	if self.value != nil{
		return fmt.Sprintf("var %s %s = %s", self.t, self.Name(), self.value.Name())
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
	blocks *list.List[*Block]

	attrs hashset.HashSet[FunctionAttribute]
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
		blocks: list.New[*Block](),
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

func (self *Function) setIndex(i uint) uint {
	if self.name != ""{
		return i
	}
	self.index = i
	return i+1
}

func (self *Function) Name()string{
	if self.name != ""{
		return "f_" + self.name
	}
	return fmt.Sprintf("f%d", self.index)
}

func (self *Function) RealName()string{
	return self.name
}

func (self *Function) Define()string{
	if self.blocks.Len() == 0{
		params := lo.Map(self.t.Params(), func(item Type, _ int) string {
			return item.String()
		})
		attrs := lo.Map(self.Attributes(), func(item FunctionAttribute, _ int) string {
			return string(item)
		})
		attrStr := stlbasic.Ternary(len(attrs)==0, "", fmt.Sprintf("\nattrs: %s", strings.Join(attrs, ";")))
		return fmt.Sprintf("%s %s(%s)%s", self.t.Ret(), self.Name(), strings.Join(params, ","), attrStr)
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
	buf.WriteString("):\n")
	if !self.attrs.Empty(){
		buf.WriteString("attrs: ")
	}
	for iter:=self.attrs.Iterator(); iter.Next(); {
		buf.WriteString(string(iter.Value()))
		if iter.HasNext(){
			buf.WriteByte(';')
		}else{
			buf.WriteByte('\n')
		}
	}
	var i uint
	for cursor:=self.blocks.Front(); cursor!=nil; cursor=cursor.Next(){
		cursor.Value.index = i
		i++
	}
	for cursor:=self.blocks.Front(); cursor!=nil; cursor=cursor.Next(){
		buf.WriteString(cursor.Value.String())
		if cursor.Next() != nil{
			buf.WriteByte('\n')
		}
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

func (self *Function) Blocks()*list.List[*Block]{
	return self.blocks
}

// FunctionAttribute 函数属性
type FunctionAttribute string

const (
	FunctionAttributeInit FunctionAttribute = "init"  // init
	FunctionAttributeFini FunctionAttribute = "fini"  // fini
	FunctionAttributeNoReturn FunctionAttribute = "noreturn"  // 函数不会返回
)

func (self *Function) SetAttribute(attr ...FunctionAttribute){
	for _, a := range attr{
		switch a {
		case FunctionAttributeInit, FunctionAttributeFini:
			if !self.t.Ret().Equal(self.ctx.Void()) || len(self.t.Params()) != 0{
				panic("unreachable")
			}
		case FunctionAttributeNoReturn:

		default:
			panic("unreachable")
		}
		self.attrs.Add(a)
	}
}

func (self *Function) ContainAttribute(attr FunctionAttribute)bool{
	return self.attrs.Contain(attr)
}

func (self *Function) Attributes()[]FunctionAttribute{
	return self.attrs.ToSlice().ToSlice()
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

func (self *Constant) setIndex(i uint) uint {
	if self.name != ""{
		return i
	}
	self.index = i
	return i+1
}

func (self *Constant) Name()string{
	if self.name != ""{
		return "c_" + self.name
	}
	return fmt.Sprintf("c%d", self.index)
}

func (self *Constant) RealName()string{
	return self.name
}

func (self *Constant) Context()*Context{
	return self.ctx
}

func (self *Constant) Define()string{
	return fmt.Sprintf("const %s %s = %s", self.Type(), self.Name(), self.value.Name())
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
