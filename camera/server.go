package camera

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
)

// Server allows serving video stream and pictures over HTTP.
type Server struct {
	ValidatePassword func(string) error
}

func getRaspividCommand(width, height, fps int) *exec.Cmd {
	return exec.Command("raspivid",
		"--nopreview", // do not show preview window (the server is headless)
		"--width", strconv.Itoa(width),
		"--height", strconv.Itoa(height),
		"--vflip", "--hflip", // flip the viewport (this brings a normal output)
		"--timeout", "0", // do not stop video streaming
		"--framerate", strconv.Itoa(fps),
		"--vstab", // enable software vertical stabilization
		"-o", "-")
}

func getRaspistillCommand(width, height, quality int) *exec.Cmd {
	return exec.Command("raspistill",
		"--nopreview",
		"--width", strconv.Itoa(width),
		"--height", strconv.Itoa(height),
		"--vflip", "--hflip",
		"--quality", strconv.Itoa(quality),
		"-o", "-")
}

type request struct {
	r *http.Request
	w http.ResponseWriter
}

func (r *request) renderError(code int, message string) {
	r.w.Header().Set("Content-Type", "text/plain")
	r.w.WriteHeader(code)
	fmt.Fprintln(r.w, message)
}

func (r *request) getIntHeader(header string, intValue *int) bool {
	value := r.r.Header.Get(header)
	if value == "" {
		return true
	}
	var e error
	if *intValue, e = strconv.Atoi(value); e != nil {
		r.renderError(http.StatusBadRequest, e.Error())
		return false
	}
	return true
}

// Handler replies to the client request for camera pictures or video.
func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	req := &request{r: r, w: w}

	if s.ValidatePassword(r.Header.Get("X-Capture-Server-PASSWORD")) != nil {
		req.renderError(http.StatusForbidden, "403 Forbidden")
		return
	}

	width := 2592
	height := 1944
	fps := 20
	quality := 80

	if filepath.Ext(r.URL.Path) == ".jpg" {
		fps = 0
	}

	if !req.getIntHeader("X-Capture-Server-QUALITY", &quality) {
		return
	}
	if !req.getIntHeader("X-Capture-Server-FPS", &fps) {
		return
	}

	if fps > 0 {
		width = 640
		height = 480
	}

	if !req.getIntHeader("X-Capture-Server-WIDTH", &width) {
		return
	}
	if !req.getIntHeader("X-Capture-Server-HEIGHT", &height) {
		return
	}

	var captureCommand *exec.Cmd

	w.Header().Set("Server", "Go (raspivid/raspistill)")
	if fps > 0 {
		w.Header().Set("Content-Type", "video/h264")
		captureCommand = getRaspividCommand(width, height, fps)
	} else {
		w.Header().Set("Content-Type", "image/jpeg")
		captureCommand = getRaspistillCommand(width, height, quality)
	}
	captureCommand.Stdout = w
	if e := captureCommand.Run(); e != nil {
		log.Printf("Error executing %s: %v", captureCommand.Path, e)
		// TODO(dotdoom): send/log stderr
		req.renderError(http.StatusInternalServerError, e.Error())
	}
}
