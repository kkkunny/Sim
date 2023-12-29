package c

import (
	"strings"

	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedhashset"
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/mir"
)

type COutputer struct {
	includes linkedhashset.LinkedHashSet[string]
	typedefs linkedhashmap.LinkedHashMap[string, pair.Pair[string, string]]
	types hashmap.HashMap[mir.Type, string]
	values hashmap.HashMap[mir.Value, string]
	blocks hashmap.HashMap[*mir.Block, string]

	structDeclBuffer strings.Builder
	varDeclBuffer strings.Builder
	varDefBuffer strings.Builder

	localVarCount uint64
}

func NewCOutputer()*COutputer{
	return &COutputer{}
}

func (self *COutputer) init(_ *mir.Module){
	self.includes = linkedhashset.NewLinkedHashSet[string]()
	self.typedefs = linkedhashmap.NewLinkedHashMap[string, pair.Pair[string, string]]()
	self.types = hashmap.NewHashMap[mir.Type, string]()
	self.values = hashmap.NewHashMap[mir.Value, string]()

	self.structDeclBuffer.Reset()
	self.varDeclBuffer.Reset()
	self.varDefBuffer.Reset()

	self.localVarCount = 0
}

func (self *COutputer) Codegen(module *mir.Module){
	self.init(module)

	for cursor:=module.Globals().Front(); cursor!=nil; cursor=cursor.Next(){
		self.codegenDeclType(cursor.Value)
	}
	for cursor:=module.Globals().Front(); cursor!=nil; cursor=cursor.Next(){
		self.codegenDefType(cursor.Value)
	}

	for cursor:=module.Globals().Front(); cursor!=nil; cursor=cursor.Next(){
		self.codegenDeclValue(cursor.Value)
	}
	for cursor:=module.Globals().Front(); cursor!=nil; cursor=cursor.Next(){
		self.codegenDefValue(cursor.Value)
	}
}

func (self *COutputer) Result()Result{
	var buf strings.Builder

	// include
	for iter:=self.includes.Iterator(); iter.Next(); {
		buf.WriteString("#include ")
		buf.WriteString(iter.Value())
		buf.WriteString(";\n")
	}
	if !self.includes.Empty(){
		buf.WriteByte('\n')
	}

	// type decl
	buf.WriteString(self.structDeclBuffer.String())
	if self.structDeclBuffer.Len() > 0{
		buf.WriteByte('\n')
	}

	// type def
	for iter:=self.typedefs.Iterator(); iter.Next(); {
		buf.WriteString(iter.Value().Second.Second)
		buf.WriteString(";\n")
	}
	if !self.typedefs.Empty(){
		buf.WriteByte('\n')
	}

	// var decl
	buf.WriteString(self.varDeclBuffer.String())
	if self.varDeclBuffer.Len() > 0{
		buf.WriteByte('\n')
	}

	// var def
	buf.WriteString(self.varDefBuffer.String())
	if self.varDefBuffer.Len() > 0{
		buf.WriteByte('\n')
	}

	return Result{
		code: buf,
	}
}
