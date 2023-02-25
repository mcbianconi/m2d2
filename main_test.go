package main

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertInline(t *testing.T) {
	testCases := []struct {
		name          string
		content       string
		expectedError bool
	}{
		{
			name:          "valid code block",
			content:       "\n```d2\na->b\n```\n",
			expectedError: false,
		},
		{
			name:          "invalid code block",
			content:       "```d2\nasasd\n```",
			expectedError: true,
		},
	}

	svgOutputPath = os.TempDir()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mdFile, _ := os.CreateTemp("", "test*.md")
			defer os.Remove(mdFile.Name())

			mdFile.Write([]byte(tc.content))

			convertInline(mdFile.Name())
			fileContent, err := os.ReadFile(mdFile.Name())
			assert.NoError(t, err)
			if tc.expectedError {
				assert.Equal(t, tc.content, string(fileContent))
			} else {
				assert.Regexp(t, regexp.MustCompile(`!\[.*\]\((.+?)\)`), string(fileContent)) // Regex pega tags no formato esperado: ![label](link)
			}
		})
	}
}

func TestGetCodeBlocks(t *testing.T) {
	mdFile, err := os.CreateTemp("", "test-codeblocks-*.md")
	assert.NoError(t, err)
	defer os.Remove(mdFile.Name())

	inline1 := "\n```d2\na->b\n```\n"
	inline2 := "\n```d2\nb->c\n```\n"
	rolha := "\n#Teste\n"
	content := inline1 + rolha + inline2
	_, err = mdFile.Write([]byte(content))
	assert.NoError(t, err)

	expectedBlocks := []InlineCode{
		{
			DiagramCode: DiagramCode{
				Content: "a->b",
				File:    mdFile,
			},
			Inline: inline1,
		},
		{
			DiagramCode: DiagramCode{
				Content: "b->c",
				File:    mdFile,
			},
			Inline: inline2,
		},
	}

	blocks := getCodeBlocks(mdFile, content)
	assert.Equal(t, expectedBlocks, blocks)
}

func TestReference(t *testing.T) {
	diagram1 := DiagramCode{
		Content: "a->b",
		File:    &os.File{},
	}

	diagram2 := DiagramCode{
		Content: "a->b",
		File:    &os.File{},
	}

	assert.Equal(t, diagram1.Content, diagram2.Content)
	assert.Equal(t, diagram1.Reference(), diagram2.Reference())
}

func TestToMarkdown(t *testing.T) {
	file, err := os.CreateTemp("", "test-markdow-*.md")
	diagramImg := DiagramImg{
		DiagramCode: DiagramCode{},
		File:        file,
	}

	assert.NoError(t, err)
	assert.Equal(t, diagramImg.ToMarkdown(), fmt.Sprintf("\n\n![%s](%s)\n\n", file.Name(), file.Name()))
}
