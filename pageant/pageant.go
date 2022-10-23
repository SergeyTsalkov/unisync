//go:build windows
// +build windows

// Taken from https://github.com/kbolino/pageant for unisync
// Package pageant provides native Go support for using PuTTY Pageant as an
// SSH agent with the golang.org/x/crypto/ssh/agent package.
// Based loosely on the Java JNA package jsch-agent-proxy-pageant.
package pageant

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	agentCopydataID = 0x804e50ba
	agentMaxMsglen  = 8192
	noError         = syscall.Errno(0)
	wmCopyData      = 0x004a
)

var (
	pageantWindowName = utf16Ptr("Pageant")
	user32            = windows.NewLazySystemDLL("user32.dll")
	findWindow        = user32.NewProc("FindWindowW")
	sendMessage       = user32.NewProc("SendMessageW")
)

// Conn is a shared-memory connection to Pageant.
// Conn implements io.Reader, io.Writer, and io.Closer.
// It is not safe to use Conn in multiple concurrent goroutines.
type Conn struct {
	window     windows.Handle
	sharedFile windows.Handle
	sharedMem  uintptr
	readOffset int
	readLimit  int
	mapName    string
}

var _ io.ReadWriteCloser = &Conn{}

// NewConn creates a new connection to Pageant.
// Ensure Close gets called on the returned Conn when it is no longer needed.
func NewConn() (*Conn, error) {
	return &Conn{}, nil
}

// Close frees resources used by Conn.
func (c *Conn) Close() error {
	if c.sharedMem == 0 {
		return nil
	}
	errUnmap := windows.UnmapViewOfFile(c.sharedMem)
	errClose := windows.CloseHandle(c.sharedFile)
	if errUnmap != nil {
		return errUnmap
	} else if errClose != nil {
		return errClose
	}
	c.sharedMem = 0
	c.sharedFile = windows.InvalidHandle
	return nil
}

func (c *Conn) Read(p []byte) (n int, err error) {
	if c.sharedMem == 0 {
		return 0, fmt.Errorf("not connected to Pageant")
	} else if c.readLimit == 0 {
		return 0, fmt.Errorf("must send request to Pageant before reading response")
	} else if c.readOffset == c.readLimit {
		return 0, io.EOF
	}
	bytesToRead := minInt(len(p), c.readLimit-c.readOffset)
	src := toSlice(c.sharedMem+uintptr(c.readOffset), bytesToRead)
	copy(p, src)
	c.readOffset += bytesToRead
	return bytesToRead, nil
}

func (c *Conn) Write(p []byte) (n int, err error) {
	if len(p) > agentMaxMsglen {
		return 0, fmt.Errorf("size of request message (%d) exceeds max length (%d)", len(p), agentMaxMsglen)
	} else if len(p) == 0 {
		return 0, fmt.Errorf("message to send is empty")
	}
	if c.sharedMem != 0 {
		err := c.Close()
		if c.sharedMem != 0 {
			return 0, fmt.Errorf("failed to close previous connection: %s", err)
		}
	}
	if err := c.establishConn(); err != nil {
		return 0, fmt.Errorf("failed to connect to Pageant: %s", err)
	}
	dst := toSlice(c.sharedMem, len(p))
	copy(dst, p)
	data := make([]byte, len(c.mapName)+1)
	copy(data, c.mapName)
	result, err := c.sendMessage(data)
	if result == 0 {
		if err != nil {
			return 0, fmt.Errorf("failed to send request to Pageant: %s", err)
		} else {
			return 0, fmt.Errorf("request refused by Pageant")
		}
	}
	messageSize := binary.BigEndian.Uint32(toSlice(c.sharedMem, 4))
	if messageSize > agentMaxMsglen-4 {
		return 0, fmt.Errorf("size of response message (%d) exceeds max length (%d)", messageSize+4, agentMaxMsglen)
	}
	c.readOffset = 0
	c.readLimit = 4 + int(messageSize)
	return len(p), nil
}

// establishConn creates a new connection to Pageant.
func (c *Conn) establishConn() error {
	window, _, err := findWindow.Call(
		uintptr(unsafe.Pointer(pageantWindowName)),
		uintptr(unsafe.Pointer(pageantWindowName)),
	)
	if window == 0 {
		if err != nil && err != noError {
			return fmt.Errorf("cannot find Pageant window: %s", err)
		} else {
			return fmt.Errorf("cannot find Pageant window, ensure Pageant is running")
		}
	}
	mapName := fmt.Sprintf("PageantRequest%08x", windows.GetCurrentThreadId())
	mapNameUTF16 := utf16Ptr(mapName)
	sharedFile, err := windows.CreateFileMapping(
		windows.InvalidHandle,
		nil,
		windows.PAGE_READWRITE,
		0,
		agentMaxMsglen,
		mapNameUTF16,
	)
	if err != nil {
		return fmt.Errorf("failed to create shared file: %s", err)
	}
	sharedMem, err := windows.MapViewOfFile(
		sharedFile,
		windows.FILE_MAP_WRITE,
		0,
		0,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to map file into shared memory: %s", err)
	}
	*c = Conn{
		window:     windows.Handle(window),
		sharedFile: sharedFile,
		sharedMem:  sharedMem,
		mapName:    mapName,
	}
	return nil
}

// sendMessage invokes user32.SendMessage to alert Pageant that data
// is available for it to read.
func (c *Conn) sendMessage(data []byte) (uintptr, error) {
	cds := copyData{
		dwData: agentCopydataID,
		cbData: uintptr(len(data)),
		lpData: uintptr(unsafe.Pointer(&data[0])),
	}
	result, _, err := sendMessage.Call(
		uintptr(c.window),
		wmCopyData,
		0,
		uintptr(unsafe.Pointer(&cds)),
	)
	if err == noError {
		return result, nil
	}
	return result, err
}

// copyData is equivalent to COPYDATASTRUCT.
// Unlike Java, Go has a native type that matches the bit width of the
// platform, so there is no need for separate 32-bit and 64-bit versions.
// Curiously, the MSDN definition of COPYDATASTRUCT says dwData is ULONG_PTR
// and cbData is DWORD, which seems to be backwards.
type copyData struct {
	dwData uint32
	cbData uintptr
	lpData uintptr
}

// minInt returns the lesser of x and y.
func minInt(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

// toSlice creates a fake slice header that allows copying to/from the block
// of memory from addr to addr+size.
func toSlice(addr uintptr, size int) []byte {
	header := reflect.SliceHeader{
		Len:  size,
		Cap:  size,
		Data: addr,
	}
	return *(*[]byte)(unsafe.Pointer(&header))
}

// utf16Ptr converts a static string not containing any zero bytes to a
// sequence of UTF-16 code units, represented as a pointer to the first one.
func utf16Ptr(s string) *uint16 {
	result, err := windows.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return result
}
