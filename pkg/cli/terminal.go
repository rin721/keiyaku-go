package cli

import "os"

// DetectInteractive 判断当前标准输入输出是否可以交互。
func DetectInteractive() bool {
	return isTerminal(os.Stdin) && isTerminal(os.Stdout)
}

func isTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}
