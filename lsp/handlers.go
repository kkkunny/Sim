package lsp

import (
	"context"
	"os"
	"strings"
	"unicode/utf16"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"

	"github.com/kkkunny/Sim/compiler/analyse"
	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/utils"
)

var log *zap.Logger

type Handler struct {
	protocol.Server
	logger *zap.Logger

	buildinPkg *hir.Package
}

func NewHandler(ctx context.Context, server protocol.Server, logger *zap.Logger) (*Handler, context.Context, error) {
	return &Handler{Server: server, logger: logger}, ctx, nil
}

func (h *Handler) Initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	pkg, err := analyse.Analyse(config.BuildInPkgPath)
	if err != nil {
		return nil, err
	}
	h.buildinPkg = pkg
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			HoverProvider:      true,
			DefinitionProvider: true,
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "sim",
			Version: "0.1.0",
		},
	}, nil
}

func (h *Handler) Hover(ctx context.Context, params *protocol.HoverParams) (result *protocol.Hover, err error) {
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.PlainText,
			Value: "hello world",
		},
	}, nil
}

func (h *Handler) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	queryWord, err := getQueryWord(params.TextDocumentPositionParams)
	if err != nil {
		return nil, err
	}

	obj, ok := h.buildinPkg.GetIdent(queryWord)
	if !ok {
		return nil, nil
	}
	globalObj, namedObj := obj.(hir.Global), obj.(utils.Named)
	filepath := globalObj.File().Path()
	fileData, err := os.ReadFile(string(filepath))
	if err != nil {
		return nil, err
	}
	fileContent := string(fileData)
	name, _ := namedObj.GetName()
	lineCount := strings.Count(fileContent[:name.Position.Begin], "\n")
	lineIndex := strings.LastIndex(fileContent[:name.Position.Begin], "\n")
	lineString := fileContent[lineIndex+1:]
	startPos := len(utf16.Encode([]rune(lineString[:name.Position.Begin-uint(lineIndex)-1])))
	return []protocol.Location{
		{
			URI: uri.File(string(filepath)),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(lineCount),
					Character: uint32(startPos),
				},
				End: protocol.Position{
					Line:      uint32(lineCount),
					Character: uint32(startPos + len(utf16.Encode([]rune(queryWord)))),
				},
			},
		},
	}, nil
}
