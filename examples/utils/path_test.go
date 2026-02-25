package utils

import (
	"fmt"
	"testing"
)

func TestIsValudPath(t *testing.T) {
	validPaths := []string{
		"/valid/path/to/文件",                   // Example Linux-like path with Chinese
		"C:/valid/path/to/文件",                 // Example Windows path with Chinese using '/'
		"C:\\valid\\path\\to\\文件",             // Example Windows path with Chinese using '\'
		"/路径/包含/符号/!@#$%^&()`-_+'=",           // Linux-like path with symbols
		"C:/路径/包含/符号/!@#$%^&()",               // Windows path with symbols using '/'
		"C:\\路径\\包含\\符号\\!@😋#$%^&()",          // Windows path with symbols using '\'
		"/valid/path/with spaces/file",        // Linux-like path with spaces
		"C:/valid/path/with spaces/file",      // Windows path with spaces using '/'
		"C:\\valid\\path\\with spaces\\file",  // Windows path with spaces using '\'
		"/C:\\valid\\path\\with spaces\\file", // Windows path with spaces using '\'
	}

	invalidPaths := []string{
		"/invalid/path/<>/to/file",          // Invalid Linux-like path
		"C:/invalid/path/<>/file",           // Invalid Windows path using '/'
		"C:\\invalid\\path\\<>\\file",       // Invalid Windows path using '\'
		"relative/path/to/file",             // Relative path (not absolute)
		"C:missing/slash/after/drive",       // Invalid Windows path missing slash after drive letter
		"/invalidpath/!@#$%^&()`-_+=:;'\"",  // Linux-like path with invalid characters
		"C:\\invalid|path\\with*characters", // Windows path with invalid characters
		"C:\\invalid|path\\with*characters", // Windows path with invalid characters
		"127.0.0.1:8080",                    // Windows path with invalid characters
	}

	fmt.Println("Valid Paths:")
	for _, path := range validPaths {
		fmt.Printf("Checking path: %s\n", path)
		if IsValidPath(path) {
			fmt.Println("Path is valid.")
		} else {
			fmt.Println("Path is invalid.")
			t.Error("Path is invalid.")
		}
	}

	fmt.Println("\nInvalid Paths:")
	for _, path := range invalidPaths {
		fmt.Printf("Checking path: %s\n", path)
		if IsValidPath(path) {
			fmt.Println("Path is valid.")
			t.Error("Path is valid.")
		} else {
			fmt.Println("Path is invalid.")
		}
	}
}
