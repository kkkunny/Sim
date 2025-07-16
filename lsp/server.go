package lsp

import (
	"context"
	"io"
	"os"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func StartServer() error {
	conn := jsonrpc2.NewConn(jsonrpc2.NewStream(&readWriteCloser{
		reader: os.Stdin,
		writer: os.Stdout,
	}))

	logger, _ := zap.NewDevelopmentConfig().Build()
	handler, ctx, err := NewHandler(
		context.Background(),
		protocol.ServerDispatcher(conn, logger),
		logger,
	)

	if err != nil {
		return err
	}

	conn.Go(ctx, protocol.ServerHandler(
		handler, jsonrpc2.MethodNotFoundHandler,
	))
	<-conn.Done()
	return nil
}

type readWriteCloser struct {
	reader io.ReadCloser
	writer io.WriteCloser
}

func (r *readWriteCloser) Read(b []byte) (int, error) {
	n, err := r.reader.Read(b)
	return n, err
}

func (r *readWriteCloser) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func (r *readWriteCloser) Close() error {
	return multierr.Append(r.reader.Close(), r.writer.Close())
}
