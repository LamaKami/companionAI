package utils

import (
	"companionAI/helper"
)

func contains(list []string, element string) bool {
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
