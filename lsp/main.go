package main

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"
	stllog "github.com/kkkunny/stl/log"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/parse"
)

func main() {
	logger := stllog.DefaultLogger(true)
	target := stlerror.MustWith(llvm.NativeTarget())

	server := lsp.NewServer(&lsp.Options{})
	server.OnDidChangeWatchedFiles(func(ctx context.Context, req *defines.DidChangeWatchedFilesParams) (err error) {
		return errors.New("123")
		for _, change := range req.Changes{
			if change.Type != defines.FileChangeTypeChanged{
				continue
			}

			path, err := stlerror.ErrorWith(filepath.Abs(string(change.Uri)))
			if err != nil{
				_ = logger.WarnError(0, err)
				continue
			}
			asts, err := parse.ParseFile(path)
			if err != nil{
				_ = logger.WarnError(0, err)
				continue
			}
			analyse.New(path, asts, target).Analyse()
		}
		return nil
	})
	server.Run()
}
