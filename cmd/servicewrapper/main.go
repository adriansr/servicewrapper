package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func quoteArgs(args []string) []string {
	res := make([]string, len(args))
	for idx, arg := range args {
		res[idx] = `"` + arg + `"`
	}
	return res
}

var errOutput = os.Stderr

func ErrMsg(format string, args... interface{}) {
	fmt.Fprintf(errOutput, format + "\n", args...)
}

var delay = flag.Uint("delay", 0, "delay before launching service (milliseconds)")
var output = flag.String("output", "", "path to file that receives service output (stdout and stderr). Paths will be relative to SYSTEM32 directory.")
var help = flag.Bool("help", false, "shows this help")
func main() {
	flag.Parse()
	if flag.NArg() < 1 || *help {
		ErrMsg("Usage: %s [flags] path/to/exe [arguments...]", os.Args[0])
		ErrMsg("\nAllowed flags:")
		flag.PrintDefaults()
		return
	}

	fileHandle := windows.Stderr
	stdHandles := []windows.Handle {windows.Stdin, windows.Stdout, windows.Stderr}

	if output != nil && *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			ErrMsg("Failed to create output file '%s': %v", *output, err)
			return
		}
		defer f.Sync()
		defer f.Close()
		ErrMsg("Writing all output to %s ...", *output)
		errOutput = f
		fileHandle = windows.Handle(f.Fd())

		// Duplicate std handles.
		processHandle := windows.CurrentProcess()
		for idx := 1; idx < 3; idx++{
			err := windows.DuplicateHandle(processHandle, fileHandle, processHandle, &stdHandles[idx], 0, true, windows.DUPLICATE_SAME_ACCESS)
			if err != nil {
				ErrMsg("Error duplicating handle #%d", idx)
				return
			}
		}

	}

	if delay != nil && *delay > 0 {
		ErrMsg("Waiting %d milliseconds before launching process...", *delay)
		windows.SleepEx(uint32(*delay), false)
	}

	si := new(windows.StartupInfo)
	si.Cb = uint32(unsafe.Sizeof(*si))
	si.Flags = windows.STARTF_USESTDHANDLES
	si.StdInput = stdHandles[0]
	si.StdOutput = stdHandles[1]
	si.StdErr = stdHandles[2]

	// Run process
	ErrMsg("Creating process ...")
	pi := new(windows.ProcessInformation)
	args := flag.Args()
	err := windows.CreateProcess(
		windows.StringToUTF16Ptr(args[0]),
		windows.StringToUTF16Ptr(strings.Join(quoteArgs(args[1:]), " ")),
		nil,
		nil,
		true,
		0,
		nil,
		nil,
		si,
		pi)
	if err != nil {
		ErrMsg("Failed to create process: %v", err)
		return
	}
	ErrMsg("Created process with PID %d", pi.ProcessId)

	// Wait until the service terminates.
	result, err := windows.WaitForSingleObject(pi.Process, windows.INFINITE)
	windows.FlushFileBuffers(fileHandle)

	if err != nil || result != windows.WAIT_OBJECT_0 {
		ErrMsg("error waiting for process termination. Code=%x Err=%v", result, err)
		return
	}
	ErrMsg("Terminated.")
}
