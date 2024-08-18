package wsinject

import (
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
)

func Test_walkDir(t *testing.T) {
	t.Run("it should visit every file ", func(t *testing.T) {
		var got []string
		want := []string{"t0", "t1", "t2"}
		tmpDir := t.TempDir()
		os.WriteFile(path.Join(tmpDir, "t0"), []byte("t0"), 0o644)
		os.WriteFile(path.Join(tmpDir, "t1"), []byte("t1"), 0o644)
		nestedDir, err := os.MkdirTemp(tmpDir, "dir0_*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		os.WriteFile(path.Join(nestedDir, "t2"), []byte("t2"), 0o644)

		wsInjectMaster(tmpDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				t.Fatalf("got err during traversal: %v", err)
			}
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
	tmpDir := t.TempDir()
	ancli.Newline = true
	fileName := "t0.html"
	os.WriteFile(path.Join(tmpDir, fileName), []byte(mockHtml), 0o777)
	nestedDir := path.Join(tmpDir, "nested")
	err := os.MkdirAll(nestedDir, 0o777)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	nestedFile := path.Join(nestedDir, "nested.html")
	os.WriteFile(nestedFile, []byte(mockHtml), 0o777)
	fs := NewFileServer(8080, "/delta-streamer-ws.js")
	_, err = fs.Setup(tmpDir)
	if err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	checkIfInjected := func(t *testing.T, filePath string) {
		b, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read mirrored file: %v", err)
		}
		if !strings.Contains(string(b), "delta-streamer.js") {
			t.Fatalf("expected mirrored file: '%v' to have been injected with content-streamer.js", filePath)
		}
	}
	t.Run("it should inject delta-streamer.js source tag", func(t *testing.T) {
		mirrorFilePath := path.Join(fs.mirrorPath, fileName)
		checkIfInjected(t, mirrorFilePath)
	})

	t.Run("it should inject delta-streamer.js souce tag to nested files", func(t *testing.T) {
		mirrorFilePath := path.Join(fs.mirrorPath, "nested", "nested.html")
		checkIfInjected(t, mirrorFilePath)
	})

	checkIfDeltaStreamerExists := func(t *testing.T, filePath string) {
		t.Helper()
		b, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to find delta-streamer.js: %v", err)
		}
		// Whatever happens in the delta streamre source code, it should mention wd-41
		if !strings.Contains(string(b), "wd-41") {
			t.Fatal("expected delta-streamer.js file to conain string 'wd-41'")
		}
	}
	t.Run("it should write the delta streamer file to root of mirror", func(t *testing.T) {
		mirrorFilePath := path.Join(fs.mirrorPath, "delta-streamer.js")
		checkIfDeltaStreamerExists(t, mirrorFilePath)
	})
}
