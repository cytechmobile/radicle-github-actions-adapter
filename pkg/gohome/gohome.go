package gohome

import "os"

func Get() (string, error) {
	return os.UserHomeDir()
}

func Expand(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := Get()
		if err != nil {
			return path
		}
		return homeDir + path[1:]
	}
	return path
}
