package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/vnFuhung2903/vcs-sms/entities"
)

type ScopeSuite struct {
	suite.Suite
}

func TestScopeSuite(t *testing.T) {
	suite.Run(t, new(ScopeSuite))
}

func (suite *ScopeSuite) TestNumberOfScope() {
	num := NumberOfScopes()
	assert.Equal(suite.T(), num, 7)
}

func (suite *ScopeSuite) TestRoleToDefaultScope() {
	scopes := UserRoleToDefaultScopes(entities.Developer, nil)
	assert.Equal(suite.T(), len(scopes), 7)
	scopes = UserRoleToDefaultScopes(entities.Manager, nil)
	assert.Equal(suite.T(), len(scopes), 4)
	scopes = UserRoleToDefaultScopes(entities.UserRole("Not-valid"), nil)
	assert.Equal(suite.T(), len(scopes), 2)
}

func (suite *ScopeSuite) TestRoleToSpecialScope() {
	scopeHashmap := int64(5)
	scopes := UserRoleToDefaultScopes(entities.Developer, &scopeHashmap)
	assert.Equal(suite.T(), len(scopes), 2)
}

func (suite *ScopeSuite) TestScopesToHashmap() {
	scopes := []string{"user:modify", "user:manager", "container:create"}
	scopeHashmap := ScopesToHashMap(scopes)
	assert.Equal(suite.T(), scopeHashmap, int64(7))
}

func (suite *ScopeSuite) TestHashmapToScopes() {
	scopeHashmap := int64(5)
	scopes := HashMapToScopes(scopeHashmap)
	assert.Equal(suite.T(), len(scopes), 2)
}
