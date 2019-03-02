package sanitize

import (
	"fmt"
	"reflect"
)

// sanitizeInt64Field sanitizes an int64 field. Requires the whole
// reflect.Value for the struct because it needs access to both the Value and
// Type of the struct.
func sanitizeInt64Field(s Sanitizer, structValue reflect.Value, idx int) error {
	fieldValue := structValue.Field(idx)
	isPtr := fieldValue.Kind() == reflect.Ptr

	tags := s.fieldTags(structValue.Type().Field(idx).Tag)

	var err error

	// Minimum value
	_, hasMin := tags["min"]
	min := int64(0)
	if hasMin {
		min, err = parseInt64(tags["min"])
		if err != nil {
			return err
		}
	}

	// Maximum value
	_, hasMax := tags["max"]
	max := int64(0)
	if hasMax {
		max, err = parseInt64(tags["max"])
		if err != nil {
			return err
		}
	}

	// Checking if minimum is not higher than maximum
	if hasMax && hasMin && max < min {
		return fmt.Errorf(
			"max less than min on int64 field '%s' during struct sanitization",
			fieldValue.Type().Name(),
		)
	}
	// Checking if minimum and maximum are above 0
	if (hasMin && min < 0) || (hasMax && max < 0) {
		return fmt.Errorf(
			"min and max on int64 field '%s' can not be below 0",
			fieldValue.Type().Name(),
		)
	}

	// Default value
	_, hasDef := tags["def"]
	def := int64(0)
	if hasDef {
		def, err = parseInt64(tags["def"])
		if err != nil {
			return err
		}

		// Making sure default is not smaller than min or higher than max
		if hasMax && def > max {
			return fmt.Errorf(
				"incompatible def and max tag components, def (%+v) is "+
					"higher than max (%+v)",
				def,
				max,
			)
		}
		if hasMin && def < min {
			return fmt.Errorf(
				"incompatible def and min tag components, def (%+v) is "+
					"lower than min (%+v)",
				def,
				min,
			)
		}
	}

	// Pointer, nil, and we have a default: set it
	if isPtr && fieldValue.IsNil() && hasDef {
		fieldValue.Set(reflect.ValueOf(&def))
		return nil
	}

	// Not nil pointer. Dereference then continue as normal
	if isPtr && !fieldValue.IsNil() {
		fieldValue = fieldValue.Elem()
	}

	// Apply min and max transforms
	if hasMin {
		oldNum := fieldValue.Int()
		if min > oldNum {
			fieldValue.SetInt(min)
		}
	}
	if hasMax {
		oldNum := fieldValue.Int()
		if max < oldNum {
			fieldValue.SetInt(max)
		}
	}

	return nil
}
