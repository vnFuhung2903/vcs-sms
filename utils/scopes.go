package utils

import (
	"slices"

	"github.com/vnFuhung2903/vcs-sms/entities"
)

var scopeHashMap = []string{"user:modify", "user:manager", "container:create", "container:view", "container:update", "container:delete", "report:mail"}

func NumberOfScopes() int {
	return len(scopeHashMap)
}

func UserRoleToDefaultScopes(role entities.UserRole, specialScopes *int64) []string {
	if specialScopes != nil {
		return HashMapToScopes(*specialScopes)
	}

	switch role {
	case entities.Developer:
		{
			return scopeHashMap
			// return []string{"user:modify", "container:create", "container:view", "container:update", "container:delete", "report:mail"}
		}
	case entities.Manager:
		{
			return []string{"user:modify", "user:manager", "container:view", "report:mail"}
		}
	default:
		{
			return []string{"user:modify", "container:view"}
		}
	}
}

func ScopesToHashMap(userScopes []string) int64 {
	res := int64(0)
	for i, scope := range scopeHashMap {
		if found := slices.Contains(userScopes, scope); found {
			res |= (1 << i)
		}
	}
	return res
}

func HashMapToScopes(scopes int64) []string {
	var userScopes []string
	for i := range scopeHashMap {
		if scopes&(1<<i) != 0 {
			userScopes = append(userScopes, scopeHashMap[i])
		}
	}
	return userScopes
}
