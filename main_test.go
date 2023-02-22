package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validD2   = "a->b"
	invalidD2 = `Regular text`
)

func asInline(d2code string) string {
	return "```d2\n" + d2code + "\n```"
}

func TestConvertIline(t *testing.T) {
	// Cria um arquivo temporário para o código d2
	mdContent := asInline(validD2)
	mdFile := mdFile(t, mdContent)

	// Chama a função ReplaceD2 com o código d2 e o diretório temporário
	err := convertInline(mdFile.Name())

	assert.Nil(t, err, "erro convertendo inline")

	newContent, err := os.ReadFile(mdFile.Name())

	assert.Nil(t, err, "arquivo de teste não criado")

	assert.True(t, isValidImageTag(string(newContent)), fmt.Sprintf("tag gerada não é válida: %s", string(newContent)))
}

func mdFile(t *testing.T, mdContent string) *os.File {
	mdFile, err := os.CreateTemp(t.TempDir(), "*.md")
	if err != nil {
		t.Fatalf("Erro ao criar arquivo temporário: %v", err)
	}
	if _, err := mdFile.WriteString(mdContent); err != nil {
		t.Fatalf("Erro ao escrever conteúdo no arquivo: %v", err)
	}
	return mdFile
}

func isValidImageTag(tag string) bool {
	regex := regexp.MustCompile(`^!\[.*\]\(.*\.(png|jpg|jpeg|gif|svg)\)$`)
	return regex.MatchString(tag)
}
