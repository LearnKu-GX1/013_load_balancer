package main

import "net/http"

// ContextKey 类型作为 r.Context().Value 的 KEY
type ContextKey string

const (
	AttemptsKey ContextKey = "attempts"
	RetriesKey  ContextKey = "retries"
)

// GetAttemptsFromContext 从 http.Request.Context 中读取 Attempts
func GetAttemptsFromContext(r *http.Request) int {
	return getIntFromContext(r, AttemptsKey, 1)
}

// GetRetriesFromContext 从 http.Request.Context 中读取 Retries
func GetRetriesFromContext(r *http.Request) int {
	return getIntFromContext(r, RetriesKey, 0)
}

func getIntFromContext(r *http.Request, key ContextKey, defaultValue int) int {
	if value, ok := r.Context().Value(key).(int); ok {
		return value
	}
	return defaultValue
}
