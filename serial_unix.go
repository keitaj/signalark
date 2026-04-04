//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// port wraps an os.File for serial communication.
type port struct {
	file *os.File
}

// openPort opens a serial port with the given baud rate.
// Configures 8N1 (8 data bits, no parity, 1 stop bit), raw mode.
func openPort(name string, baudRate int) (*port, error) {
	f, err := os.OpenFile(name, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}

	fd := int(f.Fd())

	var t syscall.Termios
	if _, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		ioctlGETATTR,
		uintptr(unsafe.Pointer(&t)),
	); errno != 0 {
		f.Close()
		return nil, fmt.Errorf("tcgetattr: %w", errno)
	}

	// Raw mode
	t.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
		syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	t.Oflag &^= syscall.OPOST
	t.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	t.Cflag &^= syscall.CSIZE | syscall.PARENB
	t.Cflag |= syscall.CS8 | syscall.CLOCAL | syscall.CREAD

	if err := setBaudRate(&t, baudRate); err != nil {
		f.Close()
		return nil, err
	}

	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = 10

	if _, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		ioctlSETATTR,
		uintptr(unsafe.Pointer(&t)),
	); errno != 0 {
		f.Close()
		return nil, fmt.Errorf("tcsetattr: %w", errno)
	}

	return &port{file: f}, nil
}

func (p *port) Read(buf []byte) (int, error)  { return p.file.Read(buf) }
func (p *port) Write(buf []byte) (int, error) { return p.file.Write(buf) }
func (p *port) Close() error                  { return p.file.Close() }
