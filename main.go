package main

import (
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
	player           videoPlayer = &stubVideoPlayer{}
	workingDirectory             = "/mnt"
	guiMutex         sync.Mutex
	nextWakeUp       = time.Now()
	font             *gofont.Font
	fb               *framebuffer.Device
	selection        int
)

func main() {
	var err error
	fb, err = framebuffer.Open("/dev/fb0")
	if err != nil {
		panic(err)
	}
	defer fb.Close()

	font, err := gofont.LoadFromFile("/usr/share/fonts/truetype/roboto/Roboto-Regular.ttf")
	if err != nil {
		panic(err)
	}
	font.PixelHeight = 35

	go renderGui()

	keys := rc.OpenInput()
	for {
		key := <-keys
		guiMutex.Lock()
		wakeUpTV()
		clearTV()
		if key == rc.KeyWindows {
			files := listFilesIn(workingDirectory)
			if len(files) == 0 {
				panic("this should not happen, at least . should be in here")
			}
			if selection < 0 {
				selection = 0
			}
			if selection >= len(files) {
				selection = len(files) - 1
			}
			x, y := 0, 0
			for i, f := range files {
				if i == selection {
					font.R, font.G, font.B = 255, 64, 255
				} else if f.isDir {
					font.R, font.G, font.B = 255, 0, 0
				} else {
					font.R, font.G, font.B = 0, 255, 0
				}
				x, y = font.Write(f.path+"\n", fb, x, y)
			}
		}
		guiMutex.Unlock()
	}
}

func renderGui() {

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
