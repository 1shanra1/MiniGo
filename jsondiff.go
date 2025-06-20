package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type UserInfo struct {
	Id    int
	Email string
}

type UserList struct {
	Users []UserInfo
}

func errCheck(e error) {
	if e != nil {
		panic(e)
	}
}

func constructMap(ul UserList) map[int]string {
	m := make(map[int]string)
	for _, user := range ul.Users {
		m[user.Id] = user.Email
	}

	return m
}

func diffCheck(old, new []byte) []int {
	var oldUsers, newUsers UserList
	var deletedList, addedList, modifiedList []int // store the index of the modified field this way.

	oldErr := json.Unmarshal(old, &oldUsers)
	newErr := json.Unmarshal(new, &newUsers)

	errCheck(oldErr)
	errCheck(newErr)

	oldMap := constructMap(oldUsers)
	newMap := constructMap(newUsers)

	for oldId, oldEmail := range oldMap {
		newEmail, ok := newMap[oldId]
		if ok {
			if !(newEmail == oldEmail) {
				modifiedList = append(modifiedList, oldId)
			}
		} else {
			deletedList = append(deletedList, oldId)
		}
	}

	for newId := range newMap {
		_, ok := oldMap[newId]
		if !ok {
			addedList = append(addedList, newId)
		}
	}

	return addedList
}

type Change struct {
	Path     string
	OldValue any
	NewValue any
}

func recursiveDiffCheck(old, new any, added, modified, deleted *[]Change, currentPath []string) error {
	switch t := old.(type) {
	case map[string]any:
		nt, ok := new.(map[string]any)
		if !ok {
			fullPath := strings.Join(currentPath, ".") // Use '.' as a separator
			// Then, append this single fullPath string to the dereferenced slice
			*modified = append(*modified, Change{Path: fullPath, OldValue: t, NewValue: nt})
		} else {
			for oldKey, oldValue := range t {
				newPath := append(currentPath, oldKey)
				newValue, newKeyExists := nt[oldKey]
				if newKeyExists {
					if !reflect.DeepEqual(oldValue, newValue) {
						recursiveDiffCheck(oldValue, newValue, added, modified, deleted, newPath)
					}
				} else {
					fullPath := strings.Join(newPath, ".")
					*deleted = append(*deleted, Change{Path: fullPath, OldValue: oldValue, NewValue: newValue})
				}
			}

			// now check for added keys
			for newKey, newValue := range nt {
				_, oldKeyExists := t[newKey]
				if !oldKeyExists {
					newPath := append(currentPath, newKey)
					fullPath := strings.Join(newPath, ".") // Use '.' as a separator
					*added = append(*added, Change{Path: fullPath, OldValue: nil, NewValue: newValue})
				}
			}
		}
	case []any:
		newValue, newValueIsSlice := new.([]any)
		if !newValueIsSlice {
			fullPath := strings.Join(currentPath, ".")
			*modified = append(*modified, Change{Path: fullPath, OldValue: t, NewValue: newValue})
		} else {
			if len(t) > len(newValue) {
				for newSliceIndex, newSliceValue := range newValue {
					if !reflect.DeepEqual(newSliceValue, t[newSliceIndex]) {
						newPath := append(currentPath, fmt.Sprintf("[%d]", newSliceIndex))
						recursiveDiffCheck(t[newSliceIndex], newSliceValue, added, modified, deleted, newPath)
					}
				}
				// What about the rest?
				for i := len(newValue); i < len(t); i++ {
					newPath := append(currentPath, fmt.Sprintf("[%d]", i))
					finalPath := strings.Join(newPath, ".")
					*deleted = append(*deleted, Change{Path: finalPath, OldValue: t[i], NewValue: nil})
				}
			}
		}
	case string:
		return nil
	case bool:
		return nil
	case float64:
		return nil
	case nil:
		return nil
	}

	return nil
}

func main() {
	var oldJson, newJson any // using generic interfaces
	old_json_data, _ := os.ReadFile("old.json")
	new_json_data, _ := os.ReadFile("new.json")

	json.Unmarshal(old_json_data, &oldJson)
	json.Unmarshal(new_json_data, newJson)

	fmt.Printf("%v", oldJson)
}
