package hiweb

// IsJSONBody 判断body是否是json，只适用于web提交
func IsJSONBody(boey []byte) bool {
	if len(boey) > 2 && (boey[0] == '{' || boey[0] == '[') {
		return true
	}
	return false
}
