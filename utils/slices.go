package utils

import (
	"companionAI/helper"
)

func contains[V int64 | float64 | string](list []V, element V) bool {
	for i := range list {
		if list[i] == element {
			return true
		}
	}
	return false
}

func RemoveElementsFromSlice(originalList []helper.EntityDataPoint, elementsToRemove []string) []helper.EntityDataPoint {
	var newList []helper.EntityDataPoint

	for _, element := range originalList {
		if !contains(elementsToRemove, element.Id) {
			newList = append(newList, element)
		}
	}

	return newList
}
