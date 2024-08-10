package wsinject

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Fileserver struct {
	origPath   string
	mirrorPath string
}

const deltaStreamer = `<script type="module" src="delta-streamer.js"></script>`

func NewFileServer() *Fileserver {
	mirrorDir := path.Join(os.TempDir(), "wd-40")
	os.Mkdir(mirrorDir, 0755)
	return &Fileserver{
		mirrorPath: mirrorDir,
	}
}

func (fs *Fileserver) mirrorMaker(p string, info os.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	fileB, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("failed to read file on path: '%v', err: %v", p, err)
	}
	injectedBytes, err := injectWebsocketScript(fileB)
	mirroredPath := path.Join(fs.mirrorPath, info.Name())
	err = os.WriteFile(mirroredPath, injectedBytes, 0755)
	if err != nil {
		return fmt.Errorf("failed to write mirrored file: %w", err)
	}
	return nil
}

func (fs *Fileserver) Setup(origPath string) (string, error) {
	walkDir(origPath, fs.mirrorMaker)
	return fs.mirrorPath, nil
}

func walkDir(root string, do func(path string, d fs.DirEntry, err error) error) error {
	err := filepath.WalkDir(root, do)
	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", root, err)
	}
	return nil
}

func injectScript(html []byte, scriptTag string) ([]byte, error) {
	htmlStr := string(html)

	// Find the location of the closing `</header>` tag
	idx := strings.Index(htmlStr, "</head>")
	if idx == -1 {
		return nil, fmt.Errorf("no </head> tag found in the HTML")
	}

	var buf bytes.Buffer

	// Write the HTML up to the closing `</head>` tag
	_, err := buf.WriteString(htmlStr[:idx])
	if err != nil {
		return nil, fmt.Errorf("failed to write pre: %w", err)
	}

	_, err = buf.WriteString(scriptTag)

	if err != nil {
		return nil, fmt.Errorf("failed to write script tag: %w", err)
	}

	_, err = buf.WriteString(htmlStr[idx:])

	if err != nil {
		return nil, fmt.Errorf("failed to write post: %w", err)
	}

	return buf.Bytes(), nil
}

func injectWebsocketScript(b []byte) ([]byte, error) {
	contentType := http.DetectContentType(b)
	// Only act on html files
	if !strings.Contains(contentType, "text/html") {
		return b, nil
	}
	b, err := injectScript(b, deltaStreamer)
	if err != nil {
		return nil, fmt.Errorf("failed to inject script tag: %w", err)
	}

	return b, nil
}
