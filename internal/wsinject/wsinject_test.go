package wsinject

import (
	"os"
	"path"
	"slices"
	"strings"
	"testing"
)

func Test_walkDir(t *testing.T) {
	t.Run("it should visit every file ", func(t *testing.T) {
		var got []string
		want := []string{"t0", "t1", "t2"}
		tmpDir := t.TempDir()
		os.WriteFile(path.Join(tmpDir, "t0"), []byte("t0"), 0644)
		os.WriteFile(path.Join(tmpDir, "t1"), []byte("t1"), 0644)
		nestedDir, err := os.MkdirTemp(tmpDir, "dir0_*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		os.WriteFile(path.Join(nestedDir, "t2"), []byte("t2"), 0644)

		walkDir(tmpDir, func(path string, d os.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}
			got = append(got, string(b))
			return nil

		})

		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatalf("expected: %v, got: %v", want, got)
		}

	})
}

const mockHtml = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title></title>
    <link href="css/style.css" rel="stylesheet">
  </head>
  <body>
  
  </body>
</html>`

func Test_Setup(t *testing.T) {
	t.Run("it should inject delta-streamer.js", func(t *testing.T) {
		tmpDir := t.TempDir()
		fileName := "t0.html"
		os.WriteFile(path.Join(tmpDir, fileName), []byte(mockHtml), 0644)
		fs := NewFileServer()
		_, err := fs.Setup(tmpDir)
		if err != nil {
			t.Fatalf("failed to setup: %v", err)
		}

		mirrorFilePath := path.Join(fs.mirrorPath, fileName)
		b, err := os.ReadFile(mirrorFilePath)
		if err != nil {
			t.Fatalf("failed to read mirrored file: %v", err)
		}
		if !strings.Contains(string(b), "delta-streamer.js") {
			t.Fatalf("expected mirrored file: '%v' to have been injected with content-streamer.js", mirrorFilePath)
		}
	})
}
