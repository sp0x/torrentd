package indexer

import "fmt"

func firstString(obj interface{}) string {
	switch typedObj := obj.(type) {
	case string:
		return typedObj
	case []string:
		if len(typedObj) == 0 {
			return typedObj[0]
		} else {
			return ""
		}
	default:
		return fmt.Sprintf("%v", obj)
	}
}
