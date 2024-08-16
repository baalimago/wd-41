package wsinject

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
	"github.com/fsnotify/fsnotify"
)

type Fileserver struct {
	masterPath string
	mirrorPath string
	wsPort     int
	wsPath     string
	watcher    *fsnotify.Watcher

	pageReloadChan chan string
}

const deltaStreamer = `<!-- This script has been injected by wd-40 and allows hot reloads -->
<script type="module" src="delta-streamer.js"></script>`

func NewFileServer(wsPort int, wsPath string) *Fileserver {
	mirrorDir, err := os.MkdirTemp("", "wd-40_*")
	if err != nil {
		panic(err)
	}
	return &Fileserver{
		mirrorPath:     mirrorDir,
		wsPort:         wsPort,
		wsPath:         wsPath,
		pageReloadChan: make(chan string),
	}
}

func (fs *Fileserver) mirrorFile(origPath, relativeName string) error {
	fileB, err := os.ReadFile(origPath)
	if err != nil {
		return fmt.Errorf("failed to read file on path: '%v', err: %v", origPath, err)
	}
	injected, injectedBytes, err := injectWebsocketScript(fileB)
	if err != nil {
		return fmt.Errorf("failed to ineject websocket script: %v", err)
	}
	if injected {
		ancli.PrintfNotice("injected delta-streamer script loading tag in: '%v'", origPath)
	}
	mirroredPath := path.Join(fs.mirrorPath, relativeName)
	err = os.WriteFile(mirroredPath, injectedBytes, 0755)
	if err != nil {
		return fmt.Errorf("failed to write mirrored file: %w", err)
	}
	return nil
}

func (fs *Fileserver) mirrorMaker(p string, info os.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}

	return fs.mirrorFile(p, info.Name())
}

func (fs *Fileserver) writeDeltaStreamerScript() error {
	err := os.WriteFile(path.Join(fs.mirrorPath, "delta-streamer.js"), []byte(fmt.Sprintf(DeltaStreamerSourceCode, fs.wsPort, fs.wsPath)), 0755)
	if err != nil {
		return fmt.Errorf("failed to write delta-streamer.js: %w", err)
	}
	return nil
}

func (fs *Fileserver) Setup(pathToMaster string) (string, error) {
	ancli.PrintfNotice("mirroring root: '%v'", pathToMaster)
	err := wsInjectMaster(pathToMaster, fs.mirrorMaker)
	if err != nil {
		return "", fmt.Errorf("failed to create websocket injected mirror: %v", err)
	}
	err = fs.writeDeltaStreamerScript()
	if err != nil {
		return "", fmt.Errorf("failed to write delta streamer file: %w", err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return "", fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}
	err = watcher.Add(pathToMaster)
	if err != nil {
		return "", fmt.Errorf("failed to add path: '%v' to watcher, err: %v", pathToMaster, err)
	}
	fs.watcher = watcher
	fs.masterPath = pathToMaster
	return fs.mirrorPath, nil
}

// Start listening to file events, update mirror and stream notifications
// on which files to update
func (fs *Fileserver) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case fsEv, ok := <-fs.watcher.Events:
			if !ok {
				return errors.New("fsnotify watcher event channel closed")
			}
			fs.handleFileEvent(fsEv)
		case fsErr, ok := <-fs.watcher.Errors:
			if !ok {
				return errors.New("fsnotify watcher error channel closed")
			}
			return fsErr
		}
	}
}

func (fs *Fileserver) notifyPageUpdate(fileName string) {
	// Make filename relative idempotently
	fs.pageReloadChan <- strings.Replace(fileName, fs.masterPath, "", -1)
}

func (fs *Fileserver) handleFileEvent(fsEv fsnotify.Event) {
	if fsEv.Has(fsnotify.Write) {
		ancli.PrintfNotice("noticed file write in orig file: '%v',", fsEv.Name)
		fs.mirrorFile(fsEv.Name, strings.Replace(fsEv.Name, fs.masterPath, "", -1))
		fs.notifyPageUpdate(fsEv.Name)
	}
}

func wsInjectMaster(root string, do func(path string, d fs.DirEntry, err error) error) error {
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

func injectWebsocketScript(b []byte) (bool, []byte, error) {
	contentType := http.DetectContentType(b)
	// Only act on html files
	if !strings.Contains(contentType, "text/html") {
		return false, b, nil
	}
	b, err := injectScript(b, deltaStreamer)
	if err != nil {
		return false, nil, fmt.Errorf("failed to inject script tag: %w", err)
	}

	return true, b, nil
}
