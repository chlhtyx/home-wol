package common

import "errors"

var (
	// ErrConfigNotLoaded 配置未加载错误
	ErrConfigNotLoaded = errors.New("配置未加载")
	// ErrDeviceNotFound 设备未找到错误
	ErrDeviceNotFound = errors.New("设备未找到")
)
