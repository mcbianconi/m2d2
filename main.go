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

type DiagramCode struct {
	Content string
	File    *os.File
}

func (c DiagramCode) Reference() string {
	hash := md5.New()
	return base64.RawURLEncoding.EncodeToString(hash.Sum([]byte(c.Content)))
}

type InlineCode struct {
	DiagramCode
	Inline string
}

type DiagramImg struct {
	DiagramCode DiagramCode
	File        *os.File
}

func (i DiagramImg) ToMarkdown() string {
	return fmt.Sprintf("![%s](%s)", i.File.Name(), i.File.Name())
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
	mdFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer mdFile.Close()

	markdown, err := os.ReadFile(mdFile.Name())
	if err != nil {
		return err
	}

	blocks := getCodeBlocks(mdFile, string(markdown))

	var updatedContent []byte
	for _, block := range blocks {
		image, err := diagramToImg(block.DiagramCode)
		if err != nil {
			log.Printf("Erro ao converter bloco de código: %v", err)
			continue
		}
		newContent := []byte(image.ToMarkdown())
		oldContent := []byte(block.Inline)
		updatedContent = bytes.Replace(markdown, oldContent, newContent, -1)
	}

	if updatedContent != nil {
		os.WriteFile(mdFile.Name(), updatedContent, 0755)
	}

	return nil
}

func getCodeBlocks(file *os.File, content string) []InlineCode {
	d2Regex := regexp.MustCompile("(?s)\n```d2(.*?)\n```\n")
	matches := d2Regex.FindAllStringSubmatch(content, -1)

	blocks := make([]InlineCode, 0, len(matches))
	for _, match := range matches {

		code := DiagramCode{
			Content: strings.TrimSpace(match[1]),
			File:    file,
		}
		inline := InlineCode{DiagramCode: code, Inline: match[0]}
		blocks = append(blocks, inline)
	}

	return blocks
}

func diagramToImg(diagramCode DiagramCode) (DiagramImg, error) {

	svgPath := getImgPath(diagramCode)
	rendered, err := renderDiagram(diagramCode.Content, svgPath)

	if err != nil {
		return DiagramImg{}, err
	}

	image := DiagramImg{
		DiagramCode: diagramCode,
		File:        rendered,
	}

	return image, nil
}

func getImgPath(block DiagramCode) string {
	imgFileName := filepath.Join(".diagrams", string(block.Reference())+".svg")
	return imgFileName
}

func renderDiagram(code string, svgPath string) (*os.File, error) {
	f := bytes.NewBufferString(code)

	cmd := exec.Command("d2", "-", svgPath)
	cmd.Stdin = f

	var out bytes.Buffer
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao gerar diagrama %s: %w\nOutput do comando: %s", svgPath, err, out.String())
	}

	return os.Open(svgPath)
}
