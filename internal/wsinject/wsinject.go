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
	"sync"

	"github.com/baalimago/go_away_boilerplate/pkg/ancli"
	"github.com/fsnotify/fsnotify"
)

type Fileserver struct {
	masterPath  string
	mirrorPath  string
	forceReload bool
	expectTLS   bool
	wsPort      int
	wsPath      string
	watcher     *fsnotify.Watcher

	pageReloadChan        chan string
	wsDispatcher          sync.Map
	wsDispatcherStarted   *bool
	wsDispatcherStartedMu *sync.Mutex
}

var ErrNoHeaderTagFound = errors.New("no header tag found")

const deltaStreamer = `<!-- This script has been injected by wd-41 and allows hot reloads -->
<script type="module" src="delta-streamer.js"></script>`

func NewFileServer(wsPort int, wsPath string, forceReload, expectTLS bool) *Fileserver {
	mirrorDir, err := os.MkdirTemp("", "wd-41_*")
	if err != nil {
		panic(err)
	}
	started := false
	return &Fileserver{
		mirrorPath:            mirrorDir,
		wsPort:                wsPort,
		wsPath:                wsPath,
		expectTLS:             expectTLS,
		forceReload:           forceReload,
		pageReloadChan:        make(chan string),
		wsDispatcher:          sync.Map{},
		wsDispatcherStarted:   &started,
		wsDispatcherStartedMu: &sync.Mutex{},
	}
}

func (fs *Fileserver) mirrorFile(origPath string) error {
	relativePath := strings.ReplaceAll(origPath, fs.masterPath, "")
	fileB, err := os.ReadFile(origPath)
	if err != nil {
		return fmt.Errorf("failed to read file on path: '%v', err: %v", origPath, err)
	}
	injected, injectedBytes, err := injectWebsocketScript(fileB)
	if err != nil {
		return fmt.Errorf("failed to inject websocket script: %v", err)
	}
	if injected {
		ancli.PrintfNotice("injected delta-streamer script loading tag in: '%v'", origPath)
	}
	mirroredPath := path.Join(fs.mirrorPath, relativePath)
	relativePathDir := path.Dir(mirroredPath)
	err = os.MkdirAll(relativePathDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create relative dir: '%v', error: %v", relativePathDir, err)
	}
	err = os.WriteFile(mirroredPath, injectedBytes, 0o755)
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
		err = fs.watcher.Add(p)
		if err != nil {
			return fmt.Errorf("failed to add recursive path: %v", err)
		}
		return nil
	}

	return fs.mirrorFile(p)
}

func (fs *Fileserver) writeDeltaStreamerScript() error {
	tlsS := ""
	if fs.expectTLS {
		tlsS = "s"
	}
	err := os.WriteFile(
		path.Join(fs.mirrorPath, "delta-streamer.js"),
		[]byte(fmt.Sprintf(deltaStreamerSourceCode, tlsS, fs.wsPort, fs.wsPath, fs.forceReload)),
		0o755)
	if err != nil {
		return fmt.Errorf("failed to write delta-streamer.js: %w", err)
	}
	return nil
}

func (fs *Fileserver) Setup(pathToMaster string) (string, error) {
	ancli.PrintfNotice("mirroring root: '%v'", pathToMaster)
	fs.masterPath = pathToMaster
	watcher, err := fsnotify.NewWatcher()
	fs.watcher = watcher
	if err != nil {
		return "", fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}
	err = wsInjectMaster(pathToMaster, fs.mirrorMaker)
	if err != nil {
		return "", fmt.Errorf("failed to create websocket injected mirror: %v", err)
	}
	err = fs.writeDeltaStreamerScript()
	if err != nil {
		return "", fmt.Errorf("failed to write delta streamer file: %w", err)
	}
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
	fs.pageReloadChan <- strings.ReplaceAll(fileName, fs.masterPath, "")
}

func (fs *Fileserver) handleFileEvent(fsEv fsnotify.Event) {
	if fsEv.Has(fsnotify.Write) {
		ancli.PrintfNotice("noticed file write in orig file: '%v'", fsEv.Name)
		fs.mirrorFile(fsEv.Name)
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
		return html, ErrNoHeaderTagFound
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
	injected := false
	// Only act on html files
	if !strings.Contains(contentType, "text/html") {
		return injected, b, nil
	}
	b, err := injectScript(b, deltaStreamer)
	injected = true
	if err != nil {
		if !errors.Is(err, ErrNoHeaderTagFound) {
			return injected, nil, fmt.Errorf("failed to inject script tag: %w", err)
		} else {
			injected = false
		}
	}

	return injected, b, nil
}
