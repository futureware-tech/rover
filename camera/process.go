package camera

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// Process handles command line and errors for the OS process that captures the video / picture.
type Process struct {
	*exec.Cmd
}

var cameraArgs = []string{
	// do not show preview window (the server is headless)
	"--nopreview",
	// flip the viewport (this brings a normal output)
	"--vflip",
	"--hflip",
	// write output to stdout
	"-o", "-",
}

func newProcess(command string, args ...string) *Process {
	return &Process{
		Cmd: exec.Command(command, append(args, cameraArgs...)...),
	}
}

// NewVideoProcess configures a new Process for capturing video.
func NewVideoProcess(width, height, fps int) *Process {
	return newProcess("raspivid",
		"--width", strconv.Itoa(width),
		"--height", strconv.Itoa(height),
		"--timeout", "0", // do not stop video streaming
		"--framerate", strconv.Itoa(fps),
		"--vstab", // enable software vertical stabilization
	)
}

// NewPictureProcess configures a new Process for capturing picture(s).
func NewPictureProcess(width, height, quality int) *Process {
	return newProcess("raspistill",
		"--width", strconv.Itoa(width),
		"--height", strconv.Itoa(height),
		"--quality", strconv.Itoa(quality),
	)
}

// Run starts the process and waits until it finishes.
func (p *Process) Run() error {
	var stderrBuffer bytes.Buffer
	p.Cmd.Stderr = &stderrBuffer

	if e := p.Cmd.Run(); e != nil {
		errorStrings := []string{"Failed to render a capture: " + e.Error()}
		if exiterr, ok := e.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.Exited() && status.ExitStatus() == 70 {
					// Code 70 for raspivid and raspistill means "not enough resources"
					// which is normally caused by another capture process running in parallel.
					errorStrings = append(errorStrings,
						"Possible reason: another capture is already running.")
				} else if status.Signaled() && status.Signal() == syscall.SIGPIPE {
					// This is a bit controversial, but SIGPIPE in the most common
					// use case (webserver) means that the socket that stdout is bound to
					// is now closed (i.e. user has aborted the request). Not an error for us.
					log.Printf("%s got SIGPIPE, most likely the output has been closed", p.Cmd.Path)
					return nil
				}
			}
			stderr := stderrBuffer.String()
			if stderr != "" {
				// The program has exited unsuccessfully; stderr might be useful.
				errorStrings = append(errorStrings, "Technical details:\n"+stderr)
			}
		}
		for index, errorString := range errorStrings {
			errorStrings[index] = strings.TrimSpace(errorString)
		}
		return errors.New(strings.Join(errorStrings, "\n"))
	}
	return nil
}
