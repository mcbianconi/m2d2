package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBlockReference(t *testing.T) {
	input1 := "```d2\na->b\n```"
	input2 := "```d2\na->b\n```"

	result1 := getBlockReference(input1)
	result2 := getBlockReference(input2)
	assert.Equal(t, input1, input2)
	assert.Equal(t, result1, result2)
}

func TestGetFullSvgPath(t *testing.T) {
	input := "example.md"
	ref := "KdOYtsIfceBSFLXhtLaH9Q"
	expected := filepath.Join(".", "example_KdOYtsIfceBSFLXhtLaH9Q.svg")
	result := getFullSvgPath(input, ref)
	assert.Equal(t, expected, result)
}

func TestRenderSVG(t *testing.T) {
	// Para este teste, deixaremos a função renderSVG como está, uma vez que ela já chama um comando externo
	// Logo, para testá-la, precisaríamos de uma dependência externa, o que não é ideal para um teste unitário
	// É possível testá-la, mas isso envolveria mais complexidade e potencialmente causaria problemas de segurança
	// Por isso, deixaremos a função renderSVG sem testes unitários

	// Como alternativa, podemos testar se a função retorna um erro caso ocorra uma falha ao executar o comando externo
	err := renderDiagram("", "")
	assert.Error(t, err)
}

func TestGetRelativePath(t *testing.T) {
	dirPath := "/path/to/dir/"
	filePath := "/path/to/dir/file.md"
	expected := "file.md"
	result := getRelativePath(dirPath, filePath)
	assert.Equal(t, expected, result)
}

func TestInlineToImg(t *testing.T) {
	f, _ := os.CreateTemp("", "m2d2-*")
	defer os.Remove(f.Name())
	blockContent := "```d2\na->b\n```"
	f.WriteString(blockContent)
	ref := getBlockReference(blockContent)
	block := CodeBlock{
		Path:     f.Name(),
		Content:  blockContent,
		FileName: getFileName(f.Name()),
	}
	// expected := fmt.Sprintf("![%s](./example_%s.svg)", ref, ref)
	expected := Image{
		Reference: ref,
		Path:      f.Name(),
	}
	result, err := diagramToImg(block)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}
