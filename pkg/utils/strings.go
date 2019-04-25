package utils

import "strings"

func PartitionString(input, separator string) (string, string, string) {
	idx := strings.Index(input, separator)
	return partitionString(input, separator, idx)
}

/*
func RPartitionString(input, separator string) (string, string, string) {
	idx := strings.LastIndex(input, separator)
	return partitionString(input, separator, idx)
}
*/

func partitionString(input, separator string, index int) (string, string, string) {
	if index < 0 {
		return input, "", ""
	}

	return input[:index], separator, input[index+len(separator):]
}
