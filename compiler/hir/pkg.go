package hir

import (
	"os"
	"path/filepath"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"

	"github.com/kkkunny/Sim/config"
)

var (
	// OfficialPackage 官方包地址
	OfficialPackage = stlerror.MustWith(NewPackage(config.ROOT))
	// StdPackage 标准库包
	StdPackage = stlerror.MustWith(OfficialPackage.GetSon("std"))
	// BuildInPackage buildin包
	BuildInPackage = stlerror.MustWith(StdPackage.GetSon("buildin"))
)

// Package 包
type Package stlos.FilePath

func NewPackage(path stlos.FilePath)(Package, stlerror.Error){
	path, err := stlerror.ErrorWith(path.Abs())
	if err != nil{
		return "", err
	}
	info, err := stlerror.ErrorWith(os.Stat(path.String()))
	if err != nil{
		return "", err
	}
	if !info.IsDir(){
		return "", stlerror.Errorf("expect a directory")
	}
	return Package(path), nil
}

// GetPackageName 获取包名
func (self Package) GetPackageName()string{
	return stlos.FilePath(self).Base()
}

// GetSon 获取子包
func (self Package) GetSon(name ...string)(Package, error){
	return NewPackage(stlos.FilePath(self).Join(name...))
}

func (self Package) Equal(dst Package)bool{
	return string(self) == string(dst)
}

func (self Package) Hash()uint64{
	return stlbasic.Hash(string(self))
}

func (self Package) String()string{
	relpath := stlerror.MustWith(stlos.FilePath(self).Rel(stlos.FilePath(OfficialPackage)))
	dirnames := strings.Split(relpath.String(), string([]rune{filepath.Separator}))
	return strings.Join(dirnames, "::")
}

func (self Package) Path() stlos.FilePath{
	return stlos.FilePath(self)
}
