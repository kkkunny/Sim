package main

import (
	"context"
	"errors"
	"log"

	"github.com/TobiasYin/go-lsp/logs"
	"github.com/TobiasYin/go-lsp/lsp"
	"github.com/TobiasYin/go-lsp/lsp/defines"
)

func main() {
	logs.Init(log.Default())
	server := lsp.NewServer(&lsp.Options{})
	server.OnDidChangeWatchedFiles(func(ctx context.Context, req *defines.DidChangeWatchedFilesParams) (err error) {
		return errors.New("123")
	})
	server.Run()
}
