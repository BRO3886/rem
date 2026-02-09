//go:build darwin

// Package eventkit provides direct access to macOS Reminders via EventKit framework.
// It uses cgo to call Objective-C code that wraps EventKit, producing a single binary
// with no external helper process needed.
//
// This package is macOS-only (darwin). On other platforms, it will not compile.
package eventkit

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework EventKit -framework Foundation
#include "eventkit_darwin.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// FetchLists returns a JSON string containing all reminder lists.
func FetchLists() (string, error) {
	cstr := C.ek_fetch_lists()
	if cstr == nil {
		return "", getError("failed to fetch lists")
	}
	defer C.ek_free(cstr)
	return C.GoString(cstr), nil
}

// FetchReminders returns a JSON string of reminders matching the given filters.
// Any filter parameter can be empty to skip that filter.
func FetchReminders(listName, completedFilter, searchQuery, dueBefore, dueAfter string) (string, error) {
	var cListName, cCompleted, cSearch, cBefore, cAfter *C.char

	if listName != "" {
		cListName = C.CString(listName)
		defer C.free(unsafe.Pointer(cListName))
	}
	if completedFilter != "" {
		cCompleted = C.CString(completedFilter)
		defer C.free(unsafe.Pointer(cCompleted))
	}
	if searchQuery != "" {
		cSearch = C.CString(searchQuery)
		defer C.free(unsafe.Pointer(cSearch))
	}
	if dueBefore != "" {
		cBefore = C.CString(dueBefore)
		defer C.free(unsafe.Pointer(cBefore))
	}
	if dueAfter != "" {
		cAfter = C.CString(dueAfter)
		defer C.free(unsafe.Pointer(cAfter))
	}

	cstr := C.ek_fetch_reminders(cListName, cCompleted, cSearch, cBefore, cAfter)
	if cstr == nil {
		return "", getError("failed to fetch reminders")
	}
	defer C.ek_free(cstr)
	return C.GoString(cstr), nil
}

// GetReminder returns a JSON string for a single reminder by ID or ID prefix.
func GetReminder(targetID string) (string, error) {
	cID := C.CString(targetID)
	defer C.free(unsafe.Pointer(cID))

	cstr := C.ek_get_reminder(cID)
	if cstr == nil {
		return "", getError("reminder not found: " + targetID)
	}
	defer C.ek_free(cstr)
	return C.GoString(cstr), nil
}

func getError(fallback string) error {
	cerr := C.ek_last_error()
	if cerr != nil {
		return fmt.Errorf("%s", C.GoString(cerr))
	}
	return fmt.Errorf("%s", fallback)
}
