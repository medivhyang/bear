package bear

import "time"

func StringOrDefault(value *string, defaultValue ...string) string {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

func IntOrDefault(value *int, defaultValue ...int) int {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func Int8OrDefault(value *int8, defaultValue ...int8) int8 {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func Int16OrDefault(value *int16, defaultValue ...int16) int16 {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func Int32OrDefault(value *int32, defaultValue ...int32) int32 {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func Int64OrDefault(value *int64, defaultValue ...int64) int64 {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func ByteOrDefault(value *byte, defaultValue ...byte) byte {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func RuneOrDefault(value *rune, defaultValue ...rune) rune {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func Float32OrDefault(value *float32, defaultValue ...float32) float32 {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func Float64OrDefault(value *float64, defaultValue ...float64) float64 {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

func BoolOrDefault(value *bool, defaultValue ...bool) bool {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

func TimeOrDefault(value *time.Time, defaultValue ...time.Time) time.Time {
	if value != nil {
		return *value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return time.Time{}
}
