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

type CodeBlock struct {
	Path     string
	Content  string
	FileName string
}

type Image struct {
	Reference string
	Path      string
}

func main() {
	dirPath := flag.String("dir_path", ".", "Caminho do diretório a ser analisado")
	flag.Parse()

	stat, err := os.Stat(*dirPath)
	if os.IsNotExist(err) || !stat.IsDir() {
		log.Fatalf("Diretório inválido: %s não é um diretório válido.", *dirPath)
	}

	err = filepath.Walk(*dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".md" {
			log.Printf("Analisando arquivo %s...", path)

			err = convertInline(path)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func convertInline(filePath string) error {
	markdownContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	blocks := getCodeBlocks(string(markdownContent), filePath)

	for _, block := range blocks {
		image, err := diagramToImg(block)
		if err != nil {
			log.Printf("Erro ao converter bloco de código: %v", err)
			continue
		}

		newContent := fmt.Sprintf("![%s](%s)", image.Reference, image.Path)
		markdownContent = bytes.Replace(markdownContent, []byte(block.Content), []byte(newContent), 1)
	}

	err = os.WriteFile(filePath, markdownContent, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getCodeBlocks(content string, path string) []CodeBlock {
	d2Regex := regexp.MustCompile("(?s)```d2(.*?)```")

	matches := d2Regex.FindAllStringSubmatch(content, -1)

	blocks := make([]CodeBlock, 0, len(matches))
	for _, match := range matches {
		block := CodeBlock{
			Path:     path,
			Content:  match[0],
			FileName: getFileName(path),
		}

		blocks = append(blocks, block)
	}

	return blocks
}

func diagramToImg(block CodeBlock) (Image, error) {
	code := getCodeFromBlock(block.Content)
	if code == "" {
		return Image{}, fmt.Errorf("código D2 não encontrado no arquivo %s bloco %s", block.Path, block.Content)
	}

	referencia := getBlockReference(block.Content)

	svgPath := getFullSvgPath(block.FileName, referencia)

	if err := renderDiagram(code, svgPath); err != nil {
		return Image{}, err
	}

	image := Image{
		Reference: referencia,
		Path:      svgPath,
	}

	return image, nil
}

func getCodeFromBlock(blockContent string) string {
	d2Regex := regexp.MustCompile("(?s)```d2(.*?)```")

	match := d2Regex.FindStringSubmatch(blockContent)

	if match == nil || len(match) < 2 {
		return ""
	}

	return match[1]
}

func getFileName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func getBlockReference(block string) string {
	hash := md5.New()
	return base64.RawURLEncoding.EncodeToString(hash.Sum([]byte(block)))
}

func getFullSvgPath(filePath string, referencia string) string {
	svgFileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	return filepath.Join(filepath.Dir(filePath), svgFileName+"_"+referencia+".svg")
}

func getRelativePath(fromPath string, toPath string) string {
	relativePath, err := filepath.Rel(filepath.Dir(fromPath), toPath)
	if err != nil {
		return ""
	}
	return filepath.ToSlash(relativePath)
}

func renderDiagram(code string, svgPath string) error {
	f := bytes.NewBufferString(code)

	cmd := exec.Command("d2", "-", svgPath)
	cmd.Stdin = f

	var out bytes.Buffer
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao gerar .svg do arquivo %s: %w\nOutput do comando: %s", svgPath, err, out.String())
	}

	return nil
}
