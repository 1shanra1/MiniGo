package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Change struct {
	Path     string
	OldValue any
	NewValue any
}

func recursiveDiffCheck(old, new any, added, modified, deleted *[]Change, currentPath []string) {
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
					*deleted = append(*deleted, Change{Path: fullPath, OldValue: oldValue, NewValue: nil})
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
			if len(t) >= len(newValue) {
				for newSliceIndex, newSliceValue := range newValue {
					if !reflect.DeepEqual(newSliceValue, t[newSliceIndex]) {
						newPath := append(currentPath, fmt.Sprintf("[%d]", newSliceIndex))
						recursiveDiffCheck(t[newSliceIndex], newSliceValue, added, modified, deleted, newPath)
					}
				}
				for i := len(newValue); i < len(t); i++ {
					newPath := append(currentPath, fmt.Sprintf("[%d]", i))
					finalPath := strings.Join(newPath, ".")
					*deleted = append(*deleted, Change{Path: finalPath, OldValue: t[i], NewValue: nil})
				}
			} else {
				for oldSliceIndex, oldSliceValue := range t {
					if !reflect.DeepEqual(oldSliceValue, newValue[oldSliceIndex]) {
						newPath := append(currentPath, fmt.Sprintf("[%d]", oldSliceIndex))
						recursiveDiffCheck(oldSliceValue, newValue[oldSliceIndex], added, modified, deleted, newPath)
					}
				}
				for i := len(t); i < len(newValue); i++ {
					newPath := append(currentPath, fmt.Sprintf("[%d]", i))
					finalPath := strings.Join(newPath, ".")
					*added = append(*added, Change{Path: finalPath, OldValue: nil, NewValue: newValue[i]})
				}
			}
		}
	case string, bool, float64:
		if !reflect.DeepEqual(old, new) {
			fullPath := strings.Join(currentPath, ".")
			*modified = append(*modified, Change{Path: fullPath, OldValue: old, NewValue: new})
		}
	case nil:
		if new != nil {
			fullPath := strings.Join(currentPath, ".")
			*modified = append(*modified, Change{Path: fullPath, OldValue: nil, NewValue: new})
		}
	default:
		if !reflect.DeepEqual(old, new) {
			fullPath := strings.Join(currentPath, ".")
			*modified = append(*modified, Change{Path: fullPath, OldValue: old, NewValue: new})
		}
		fmt.Printf("Unhandled type for old value at path %s: %T\n", strings.Join(currentPath, "."), t) // Debugging
	}
}

func main() {
	// 1. Initialize the slices to hold the diff results
	var addedChanges []Change // Declared as actual slices, not nil pointers
	var modifiedChanges []Change
	var deletedChanges []Change

	// currentPath for the initial call, an empty slice is perfect for the root
	var currentPath []string

	// 2. Read JSON data from files
	oldJsonData, err := os.ReadFile("old.json")
	if err != nil {
		fmt.Printf("Error reading old.json: %v\n", err)
		os.Exit(1) // Exit if file cannot be read
	}

	newJsonData, err := os.ReadFile("new.json")
	if err != nil {
		fmt.Printf("Error reading new.json: %v\n", err)
		os.Exit(1) // Exit if file cannot be read
	}

	// 3. Unmarshal JSON data into generic interfaces
	var oldJson, newJson any // Using generic interfaces to hold parsed JSON
	err = json.Unmarshal(oldJsonData, &oldJson)
	if err != nil {
		fmt.Printf("Error unmarshaling old.json: %v\n", err)
		os.Exit(1) // Exit if JSON is invalid
	}

	err = json.Unmarshal(newJsonData, &newJson)
	if err != nil {
		fmt.Printf("Error unmarshaling new.json: %v\n", err)
		os.Exit(1) // Exit if JSON is invalid
	}

	// 4. Call the recursive diff checker
	// Pass the ADDRESS of the slices so the function can modify them directly
	fmt.Println("Performing JSON diff...")
	recursiveDiffCheck(oldJson, newJson, &addedChanges, &modifiedChanges, &deletedChanges, currentPath)
	if err != nil {
		fmt.Printf("Error during diff check: %v\n", err)
		os.Exit(1)
	}

	// 5. Print the results
	fmt.Println("\n--- Added Changes ---")
	if len(addedChanges) == 0 {
		fmt.Println("No additions found.")
	} else {
		for _, change := range addedChanges {
			fmt.Printf("  Path: %s, New Value: %+v\n", change.Path, change.NewValue)
		}
	}

	fmt.Println("\n--- Modified Changes ---")
	if len(modifiedChanges) == 0 {
		fmt.Println("No modifications found.")
	} else {
		for _, change := range modifiedChanges {
			fmt.Printf("  Path: %s, Old Value: %+v, New Value: %+v\n", change.Path, change.OldValue, change.NewValue)
		}
	}

	fmt.Println("\n--- Deleted Changes ---")
	if len(deletedChanges) == 0 {
		fmt.Println("No deletions found.")
	} else {
		for _, change := range deletedChanges {
			fmt.Printf("  Path: %s, Old Value: %+v\n", change.Path, change.OldValue)
		}
	}

	fmt.Println("\nDiff complete.")
}
