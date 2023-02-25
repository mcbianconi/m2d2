package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

type DiagramCode struct {
	Content string
	File    *os.File
}

func (c DiagramCode) Reference() string {
	hashBytes := sha256.Sum256([]byte(c.Content))
	return hex.EncodeToString(hashBytes[:])
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
	return fmt.Sprintf("\n\n![%s](%s)\n\n", i.File.Name(), i.File.Name())
}

var svgOutputPath string

func main() {
	dirPath := flag.String("dir", ".", "Caminho do diretório a ser analisado")
	outputPath := flag.String("output-dir", ".diagrams", "Caminho do diretório que vai armazenar as imagens geradas")
	flag.Parse()

	svgOutputPath = *outputPath

	stat, err := os.Stat(*dirPath)
	if os.IsNotExist(err) || !stat.IsDir() {
		log.Fatalf("Diretório inválido: %s não é um diretório válido.", *dirPath)
	}

	stat, err = os.Stat(svgOutputPath)
	if stat == nil || os.IsNotExist(err) {
		err := os.Mkdir(svgOutputPath, 0755)
		if err != nil {
			log.Fatal(err)
		}
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
	rendered, err := renderDiagram(diagramCode, svgPath)

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
	imgFileName := filepath.Join(svgOutputPath, block.Reference()+".svg")
	return imgFileName
}

func renderDiagram(code DiagramCode, svgPath string) (*os.File, error) {
	ruler, _ := textmeasure.NewRuler()
	diagram, _, _ := d2lib.Compile(context.Background(), code.Content, &d2lib.CompileOptions{
		Layout: d2dagrelayout.DefaultLayout,
		Ruler:  ruler,
	})
	out, _ := d2svg.Render(diagram, &d2svg.RenderOpts{
		Pad: d2svg.DEFAULT_PADDING,
	})

	_, err := os.Create(svgPath)
	if err != nil {
		log.Fatalln(err)
	}
	_ = os.WriteFile(svgPath, out, 0600)

	return os.Open(svgPath)
}
