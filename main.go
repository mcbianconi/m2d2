package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Regex para buscar referências a imagens .d2 dentro de blocos de código d2
var d2Regex = regexp.MustCompile("(?s)```d2(.*?)```")

func main() {
	// Parsing dos argumentos da linha de comando
	dirPath := flag.String("dir_path", ".", "Caminho do diretório a ser analisado")
	flag.Parse()

	// Verifica se o diretório existe
	stat, err := os.Stat(*dirPath)
	if os.IsNotExist(err) || !stat.IsDir() {
		log.Fatalf("Diretório inválido: %s não é um diretório válido.", *dirPath)
	}

	scan_dir(*dirPath)
}

func scan_dir(dirPath string) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".md" {
			log.Printf("Analisando arquivo %s...", path)
			return convertInline(path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func convertInline(path string) error {

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	newContent := d2Regex.ReplaceAllStringFunc(string(content), func(match string) string {
		replaced, err := inlineToImg(match, path)
		if err != nil {
			// Se não conseguiu converter loga mas retorna o conteúdo original
			return string(content)
		}
		return replaced
	})

	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return err
	}
	return nil
}

func inlineToImg(blockContent string, filePath string) (string, error) {
	// Extrai o código d2 do bloco de código
	match := d2Regex.FindStringSubmatch(blockContent)

	if match == nil || len(match) < 2 {
		return "", fmt.Errorf("código D2 não encontrado no arquivo %s bloco %s", filePath, blockContent)
	}

	code := match[1]

	// Define o caminho para o arquivo .svg de destino
	svgFileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	referencia := getBlockReference(blockContent)
	fullSvgPath := filepath.Join(filepath.Dir(filePath), svgFileName+"_"+referencia+".svg")

	// Cria um arquivo temporário com o conteúdo de `code`
	f, err := os.CreateTemp("", "tmp-*.d2")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(code); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	// Chama a ferramenta externa para gerar o arquivo .svg a partir do código d2
	if err := renderSVG(f, fullSvgPath); err != nil {
		return "", err
	}

	newContent := fmt.Sprintf("![%s](%s)", referencia, filepath.Join(".", strings.TrimPrefix(fullSvgPath, filepath.Dir(filePath))))

	// Retorna o bloco de código com uma imagem apontando para o arquivo .svg gerado
	return newContent, nil
}

func renderSVG(f *os.File, svgPath string) error {
	cmd := exec.Command("d2", f.Name(), svgPath)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("erro ao gerar .svg do arquivo %s: %w\nOutput do comando: %s", f.Name(), err, out.String())
	}
	return nil
}

func getBlockReference(block string) string {
	hash := md5.New()
	return base64.RawURLEncoding.EncodeToString(hash.Sum([]byte(block)))
}
