//go:build !ci
// +build !ci

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework UserNotifications

#include <stdbool.h>
#include <stdlib.h>

bool isBundled2();
void sendNotification2(char *title, char *content);
*/
import "C"

import (
	"strings"
	"unsafe"

	"fyne.io/fyne/v2"
)

func SendNotification(n *fyne.Notification) {
	if C.isBundled2() {
		titleStr := C.CString(n.Title)
		defer C.free(unsafe.Pointer(titleStr))
		contentStr := C.CString(n.Content)
		defer C.free(unsafe.Pointer(contentStr))

		C.sendNotification2(titleStr, contentStr)
		return
	}
}

func escapeNotificationString(in string) string {
	noSlash := strings.ReplaceAll(in, "\\", "\\\\")
	return strings.ReplaceAll(noSlash, "\"", "\\\"")
}

/* func fallbackSend(cTitle, cContent *C.char) {
	title := C.GoString(cTitle)
	content := C.GoString(cContent)
	fmt.Println(title, content)
	// fallbackNotification(title, content)
}
*/
