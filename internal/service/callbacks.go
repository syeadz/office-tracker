package service

import "sync"

var (
	attendanceCallbackMux     sync.RWMutex
	attendanceChangeCallback  func()
	environmentCallbackMux    sync.RWMutex
	environmentChangeCallback func()
)

// SetAttendanceChangeCallback registers a callback to be invoked when attendance changes.
func SetAttendanceChangeCallback(cb func()) {
	attendanceCallbackMux.Lock()
	defer attendanceCallbackMux.Unlock()
	attendanceChangeCallback = cb
}

// TriggerAttendanceChangeCallback invokes the registered attendance-change callback (if any) asynchronously.
func TriggerAttendanceChangeCallback() {
	attendanceCallbackMux.RLock()
	cb := attendanceChangeCallback
	attendanceCallbackMux.RUnlock()
	if cb != nil {
		go cb()
	}
}

// SetEnvironmentChangeCallback registers a callback to be invoked when environmental data changes.
func SetEnvironmentChangeCallback(cb func()) {
	environmentCallbackMux.Lock()
	defer environmentCallbackMux.Unlock()
	environmentChangeCallback = cb
}

// TriggerEnvironmentChangeCallback invokes the registered environment-change callback (if any) asynchronously.
func TriggerEnvironmentChangeCallback() {
	environmentCallbackMux.RLock()
	cb := environmentChangeCallback
	environmentCallbackMux.RUnlock()
	if cb != nil {
		go cb()
	}
}
