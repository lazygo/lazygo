package utils

import (
	"regexp"
	"strings"
)

/*
FROM GPT4
isAbsolutePath 函数:

这个函数通过分别调用 isAbsoluteLinuxPath 和 isAbsoluteWindowsPath 来判断路径是否为绝对路径。
isAbsoluteLinuxPath 函数:

检查路径是否以 / 开头，以此判断其是否为 Linux 绝对路径。
isAbsoluteWindowsPath 函数:

检查路径是否以驱动器号和 :\ 或 :/ 开头，以此判断其是否为 Windows 绝对路径。
路径格式验证:

isValidLinuxPathFormat 和 isValidWindowsPathFormat 使用正则表达式分别验证 Linux 和 Windows 的路径格式。
路径规范化:

normalizeWindowsPath 函数在路径看起来像 Windows 路径时才将 / 替换为 \。
通过这种方式，我们可以确保在 Linux 服务器上正确验证路径的合法性，包括处理 Windows 路径格式，并且兼容用户错误地使用 Windows 路径分隔符 / 的情况。根据具体需求，可以进一步完善路径检查逻辑。
*/

// isAbsolutePath checks if the path is an absolute path for both Linux and Windows.
func isAbsolutePath(path string) bool {
	return isAbsoluteLinuxPath(path) || isAbsoluteWindowsPath(path)
}

// isAbsoluteLinuxPath checks if the path is an absolute path in Linux.
func isAbsoluteLinuxPath(path string) bool {
	return strings.HasPrefix(path, "/")
}

// isAbsoluteWindowsPath checks if the path is an absolute path in Windows.
func isAbsoluteWindowsPath(path string) bool {
	path = strings.TrimLeft(path, "/")
	// Windows absolute path starts with a drive letter followed by :\
	if len(path) > 2 && path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
		return true
	}
	return false
}

// isValidLinuxPathFormat validates the Linux path format using regex.
func isValidLinuxPathFormat(path string) bool {
	// Linux path regex: starts with a slash and contains valid characters including Chinese and spaces.
	regex := `^(/[^/<>:"|?*\r\n]+)+/?$`
	re := regexp.MustCompile(regex)
	return re.MatchString(path)
}

// isValidWindowsPathFormat validates the Windows path format using regex.
func isValidWindowsPathFormat(path string) bool {
	// Updated Windows path regex to allow more characters including Chinese and spaces.
	regex := `^[a-zA-Z]:\\(?:[^<>:"/\\|?*\r\n]+\\)*[^<>:"/\\|?*\r\n]*$`
	re := regexp.MustCompile(regex)
	return re.MatchString(path)
}

// normalizeWindowsPath normalizes a Windows path by replacing '/' with '\'.
func normalizeWindowsPath(path string) string {
	// Only replace '/' with '\' if the path seems like a Windows path
	if isAbsoluteWindowsPath(path) {
		path = strings.TrimLeft(path, "/")
		return strings.ReplaceAll(path, "/", "\\")
	}
	return path
}

// IsValidPath checks if the path is valid for either Linux or Windows.
func IsValidPath(path string) bool {
	if !isAbsolutePath(path) {
		return false
	}

	// Normalize the path for Windows style paths
	normalizedPath := normalizeWindowsPath(path)

	return isValidLinuxPathFormat(normalizedPath) || isValidWindowsPathFormat(normalizedPath)
}
