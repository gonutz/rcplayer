package main

import (
	"fmt"
	"github.com/gonutz/framebuffer"
	"github.com/gonutz/gofont"
	"github.com/gonutz/rc"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	player            videoPlayer = &omxplayer{}
	workingDirectory              = "/mnt"
	guiMutex          sync.Mutex
	nextWakeUp        = time.Now()
	font              *gofont.Font
	fb                *framebuffer.Device
	selection         int
	filesInWorkingDir []file
	guiDirty          bool
	zoom              = medium
)

const (
	small = iota
	medium
	large
)

func main() {
	var err error
	fb, err = framebuffer.Open("/dev/fb0")
	if err != nil {
		panic(err)
	}
	defer fb.Close()

	font, err = gofont.LoadFromFile("/usr/share/fonts/truetype/roboto/Roboto-Regular.ttf")
	if err != nil {
		panic(err)
	}

	go renderGui()

	keys := rc.OpenInput()
	for {
		key := <-keys
		guiMutex.Lock()

		if player.isRunning() {
			switch key {
			case rc.KeyStop:
				logError(player.stopVideo())
			case rc.KeyVolumeDown:
				logError(player.volumeDown())
			case rc.KeyVolumeUp:
				logError(player.volumeUp())
			case rc.KeyChapterBack:
				logError(player.back10Minutes())
			case rc.KeyChapterForward:
				logError(player.forward10Minutes())
			case rc.KeyRewind:
				logError(player.back30Seconds())
			case rc.KeyFastForward:
				logError(player.forward30Seconds())
			case rc.KeyPause:
				logError(player.playPause())
			case rc.KeyPlay:
				logError(player.playPause())
			}
		} else {
			guiDirty = true
			switch key {
			case rc.KeyWindows:
				refreshWorkingDir()
			case rc.KeyUp:
				selection--
				if selection < 0 {
					selection = len(filesInWorkingDir) - 1
				}
			case rc.KeyDown:
				selection++
				if selection >= len(filesInWorkingDir) {
					selection = 0
				}
			case rc.KeyProgramUp:
				selection -= 10
				if selection < 0 {
					selection = 0
				}
			case rc.KeyProgramDown:
				selection += 10
				if selection >= len(filesInWorkingDir) {
					selection = len(filesInWorkingDir) - 1
				}
			case rc.KeyOK:
				if filesInWorkingDir[selection].isDir {
					workingDirectory = filesInWorkingDir[selection].path
					refreshWorkingDir()
					selection = 0
				} else {
					err := player.playVideo(filesInWorkingDir[selection].path)
					if err != nil {
						fmt.Println("Cannot play", filesInWorkingDir[selection].path, err)
					}
				}
			case rc.KeyBack:
				// assumption: the first entry is the parent directory
				workingDirectory = filesInWorkingDir[0].path
				refreshWorkingDir()
			case rc.Key1:
				zoom = small
			case rc.Key2:
				zoom = medium
			case rc.Key3:
				zoom = large
			default:
				guiDirty = false
			}
		}

		guiMutex.Unlock()
	}
}

func refreshWorkingDir() {
	filesInWorkingDir = listFilesIn(workingDirectory)
	if len(filesInWorkingDir) == 0 {
		panic("this should not happen, at least . should be in here")
	}
	if selection < 0 {
		selection = 0
	}
	if selection >= len(filesInWorkingDir) {
		selection = len(filesInWorkingDir) - 1
	}
}

func regularFontSize() int {
	switch zoom {
	case small:
		return 20
	case medium:
		return 35
	case large:
		return 50
	}
	return 35
}

func selectedFontSize() int {
	switch zoom {
	case small:
		return 25
	case medium:
		return 45
	case large:
		return 75
	}
	return 45
}

func renderGui() {
	for {
		guiMutex.Lock()
		if guiDirty {
			wakeUpTV()
			clearTV()
			x, y := 0, 0
			for i, f := range filesInWorkingDir {
				font.PixelHeight = regularFontSize()
				if i == selection {
					font.PixelHeight = selectedFontSize()
					font.R, font.G, font.B = 255, 64, 255
				} else if f.isDir {
					font.R, font.G, font.B = 255, 0, 0
				} else {
					font.R, font.G, font.B = 0, 255, 0
				}
				x, y = font.Write(f.path+"\n", fb, x, y)
			}
			guiDirty = false
		}
		guiMutex.Unlock()

		time.Sleep(500 * time.Millisecond)
	}
}

func wakeUpTV() error {
	if time.Now().After(nextWakeUp) {
		tty, err := os.OpenFile("/dev/tty1", os.O_WRONLY, os.ModeDevice)
		if err != nil {
			return err
		}
		_, err = tty.Write([]byte{0x1B, 0x5B, 0x39, 0x3B, 0x30, 0x5D})
		if err == nil {
			nextWakeUp = time.Now().Add(time.Minute)
		}
		return err
	} else {
		return nil
	}
}

func clearTV() {
	draw.Draw(fb, fb.Bounds(), image.NewUniform(color.RGBA{0, 0, 0, 255}), image.ZP, draw.Src)
}

func listFilesIn(dir string) (files []file) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == dir {
			return nil
		}

		if err == nil {
			files = append(files, file{path, info.IsDir()})
		}
		if info != nil && info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	files = append(files, file{filepath.Dir(dir), true})
	sort.Sort(fileList(files))
	return
}

type file struct {
	path  string
	isDir bool
}

type fileList []file

func (f fileList) Len() int { return len(f) }

func (f fileList) Less(i, j int) bool {
	if f[i].isDir != f[j].isDir {
		if f[i].isDir {
			return true
		}
		return false
	}

	return strings.Compare(strings.ToLower(f[i].path), strings.ToLower(f[j].path)) < 0
}

func (f fileList) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
