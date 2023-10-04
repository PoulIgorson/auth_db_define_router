package base

import "strings"

func getOffset(key string) int {
	if len(key) < 2 {
		return 0
	}
	if strings.Contains("!<=>", string(key[len(key)-2])) {
		return 2
	}
	return 1
}

func checkKey(key string, compareRes int) bool {
	if len(key) < 2 {
		return compareRes == 0
	}
	switch key[len(key)-getOffset(key):] {
	case "<":
		return compareRes == -1
	case "=":
		return compareRes == 0
	case ">":
		return compareRes == 1
	case "<=":
		return compareRes == -1 || compareRes == 0
	case ">=":
		return compareRes == 1 || compareRes == 0
	case "!=":
		return compareRes != 0
	}
	return compareRes == 0
}
