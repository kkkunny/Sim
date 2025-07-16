package lsp

import (
	"os"
	"slices"
	"strings"
	"unicode/utf16"

	"go.lsp.dev/protocol"

	"github.com/kkkunny/Sim/compiler/lex"
)

func getQueryWord(param protocol.TextDocumentPositionParams) (string, error) {
	fileData, err := os.ReadFile(param.TextDocument.URI.Filename())
	if err != nil {
		return "", err
	}
	fileContent := string(fileData)
	lines := slices.Collect(strings.Lines(fileContent))
	line := lines[param.Position.Line]

	lineUTF16 := utf16.Encode([]rune(line))
	offset := len(string(utf16.Decode(lineUTF16[:param.Position.Character])))
	buf := []byte{line[offset]}
	for cursor := offset - 1; cursor >= 0; cursor-- {
		if !lex.ByteIsIdent(line[cursor]) {
			break
		}
		buf = slices.Insert(buf, 0, line[cursor])
	}
	for cursor := offset + 1; cursor < len(line); cursor++ {
		if !lex.ByteIsIdent(line[cursor]) {
			break
		}
		buf = append(buf, line[cursor])
	}
	return string(buf), nil
}
