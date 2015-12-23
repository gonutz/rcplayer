package main

import (
	"io"
	"os/exec"
)

type videoPlayer interface {
	playVideo(path string) error
	stopVideo() error
	playPause() error
	volumeDown() error
	volumeUp() error
	back30Seconds() error
	forward30Seconds() error
	back10Minutes() error
	forward10Minutes() error
	isRunning() bool
}

type omxplayer struct {
	running bool
	play    *exec.Cmd
	out     io.ReadCloser
	in      io.WriteCloser
}

func (v *omxplayer) isRunning() bool {
	return v.running
}

func (v *omxplayer) playVideo(path string) (err error) {
	err = v.stopVideo()
	if err != nil {
		return
	}
	v.play = exec.Command("omxplayer", "-wr", path)
	v.out, err = v.play.StdoutPipe()
	if err != nil {
		return
	}
	v.in, err = v.play.StdinPipe()
	if err != nil {
		return
	}
	err = v.play.Start()
	if err != nil {
		return
	}
	v.running = true
	return nil
}

func (v *omxplayer) stopVideo() error {
	err := v.writeIfRunning("q")
	v.running = false
	return err
}

func (v *omxplayer) playPause() error {
	return v.writeIfRunning(" ")
}

func (v *omxplayer) volumeDown() error {
	return v.writeIfRunning("-")
}

func (v *omxplayer) volumeUp() error {
	return v.writeIfRunning("+")
}

func (v *omxplayer) back30Seconds() error {
	// left arrow, ASCII sequence: Esc[D
	return v.writeIfRunning(string([]byte{27, 91, 68}))
}

func (v *omxplayer) forward30Seconds() error {
	// right arrow, ASCII sequence: Esc[C
	return v.writeIfRunning(string([]byte{27, 91, 67}))
}

func (v *omxplayer) back10Minutes() error {
	// down arrow, ASCII sequence: Esc[B
	return v.writeIfRunning(string([]byte{27, 91, 66}))
}

func (v *omxplayer) forward10Minutes() error {
	// up arrow, ASCII sequence: Esc[A
	return v.writeIfRunning(string([]byte{27, 91, 65}))
}

func (v *omxplayer) writeIfRunning(msg string) (err error) {
	if v.running {
		_, err = v.in.Write([]byte(msg))
	}
	return
}

type stubVideoPlayer struct {
	running bool
}

func (v *stubVideoPlayer) playVideo(path string) error {
	v.running = true
	return nil
}

func (v *stubVideoPlayer) stopVideo() error {
	v.running = false
	return nil
}

func (v *stubVideoPlayer) isRunning() bool         { return v.running }
func (v *stubVideoPlayer) playPause() error        { return nil }
func (v *stubVideoPlayer) volumeDown() error       { return nil }
func (v *stubVideoPlayer) volumeUp() error         { return nil }
func (v *stubVideoPlayer) back30Seconds() error    { return nil }
func (v *stubVideoPlayer) forward30Seconds() error { return nil }
func (v *stubVideoPlayer) back10Minutes() error    { return nil }
func (v *stubVideoPlayer) forward10Minutes() error { return nil }
