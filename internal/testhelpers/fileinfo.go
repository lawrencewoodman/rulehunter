package testhelpers

import "time"

type FileInfo struct {
	name    string
	modTime time.Time
}

func NewFileInfo(name string, modTime time.Time) FileInfo {
	return FileInfo{name: name, modTime: modTime}
}

func (f FileInfo) Name() string       { return f.name }
func (f FileInfo) ModTime() time.Time { return f.modTime }
func (f FileInfo) String() string {
	return f.name + ": " + f.modTime.String()
}
