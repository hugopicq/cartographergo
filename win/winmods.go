package win

import (
	"golang.org/x/sys/windows"
)

var (
	//TODO : Investigate LoadLibraryEx and NewLazySystemDLL
	ModKernel32 = windows.NewLazyDLL("user32.dll")

	ProcMessageBox = ModKernel32.NewProc("MessageBoxW")
)
