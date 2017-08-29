package buffer

import "errors"

// ErrClosedBufferPool 是表示缓冲池已关闭的错误的变量。
var ErrClosedBufferPool = errors.New("缓冲池已被关闭")

// ErrClosedBuffer 是表示缓冲器已关闭的错误的变量。
var ErrClosedBuffer = errors.New("缓冲器已关闭")