package config

import "reflect"

func diffEvent(old, new any) Event {
	var changedKeys []string

	if old == nil || new == nil {
		return Event{
			ChangedKeys: changedKeys,
			OldConfig:   old,
			NewConfig:   new,
		}
	}

	oldVal := reflect.ValueOf(old)
	newVal := reflect.ValueOf(new)

	// Dereference pointers if needed
	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	// Only compute diffs for structs
	if oldVal.Kind() == reflect.Struct && newVal.Kind() == reflect.Struct {
		oldType := oldVal.Type()
		for i := 0; i < oldVal.NumField(); i++ {
			field := oldType.Field(i)
			oldFieldVal := oldVal.Field(i)
			newFieldVal := newVal.Field(i)

			if !reflect.DeepEqual(oldFieldVal.Interface(), newFieldVal.Interface()) {
				changedKeys = append(changedKeys, field.Name)
			}
		}
	}

	return Event{
		ChangedKeys: changedKeys,
		OldConfig:   old,
		NewConfig:   new,
	}
}
