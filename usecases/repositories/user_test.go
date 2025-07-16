package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vnFuhung2903/vcs-sms/entities"
)

type UserRepoSuite struct {
	suite.Suite
	db   *gorm.DB
	repo IUserRepository
}

func (suite *UserRepoSuite) SetupTest() {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(suite.T(), err)
	err = gormDB.AutoMigrate(&entities.User{})
	assert.NoError(suite.T(), err)
	suite.db = gormDB
	suite.repo = NewUserRepository(gormDB)
}

func (suite *UserRepoSuite) TearDownTest() {
	sqlDB, err := suite.db.DB()
	assert.NoError(suite.T(), err)
	sqlDB.Close()
}

func TestUserRepoSuite(t *testing.T) {
	suite.Run(t, new(UserRepoSuite))
}

func (suite *UserRepoSuite) TestCreateAndFindById() {
	user, err := suite.repo.Create("alice", "hash123", "alice@example.com", entities.Manager, 3)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)

	found, err := suite.repo.FindById(user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "alice", found.Username)
}

func (suite *UserRepoSuite) TestCreateDuplicateEmail() {
	_, err := suite.repo.Create("dave", "pass", "dave@example.com", entities.Developer, 2)
	assert.NoError(suite.T(), err)

	_, err = suite.repo.Create("dave2", "pass", "dave@example.com", entities.Developer, 2)
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestFindByIdNotFound() {
	_, err := suite.repo.FindById("non-existent-id")
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestFindByName() {
	_, err := suite.repo.Create("bob", "pass", "bob@example.com", entities.Developer, 1)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByName("bob")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "bob", found.Username)
}

func (suite *UserRepoSuite) TestFindByNameNotFound() {
	_, err := suite.repo.FindByName("unknown")
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestFindByEmail() {
	_, err := suite.repo.Create("carol", "secret", "carol@example.com", entities.Developer, 2)
	assert.NoError(suite.T(), err)

	found, err := suite.repo.FindByEmail("carol@example.com")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "carol", found.Username)
}

func (suite *UserRepoSuite) TestFindByEmailNotFound() {
	_, err := suite.repo.FindByEmail("notfound@example.com")
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestUpdatePassword() {
	user, _ := suite.repo.Create("eve", "oldpass", "eve@example.com", entities.Developer, 0)
	err := suite.repo.UpdatePassword(user, "newhash")
	assert.NoError(suite.T(), err)

	updated, _ := suite.repo.FindById(user.ID)
	assert.Equal(suite.T(), "newhash", updated.Hash)
}

func (suite *UserRepoSuite) TestUpdateRole() {
	user, _ := suite.repo.Create("frank", "hash", "frank@example.com", entities.Developer, 0)
	err := suite.repo.UpdateRole(user, entities.Manager)
	assert.NoError(suite.T(), err)

	updated, _ := suite.repo.FindById(user.ID)
	assert.Equal(suite.T(), entities.Manager, updated.Role)
}

func (suite *UserRepoSuite) TestUpdateScope() {
	user, _ := suite.repo.Create("grace", "hash", "grace@example.com", entities.Developer, 1)
	err := suite.repo.UpdateScope(user, 5)
	assert.NoError(suite.T(), err)

	updated, _ := suite.repo.FindById(user.ID)
	assert.Equal(suite.T(), int64(5), updated.Scopes)
}

func (suite *UserRepoSuite) TestUpdateNilUser() {
	err := suite.repo.UpdateRole(nil, entities.Manager)
	assert.Error(suite.T(), err)

	err = suite.repo.UpdateRole(nil, entities.Manager)
	assert.Error(suite.T(), err)

	err = suite.repo.UpdateScope(nil, 1)
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestDelete() {
	user, _ := suite.repo.Create("heidi", "hash", "heidi@example.com", entities.Developer, 1)
	err := suite.repo.Delete(user.ID)
	assert.NoError(suite.T(), err)

	_, err = suite.repo.FindById("heidi")
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestDeleteNonExistent() {
	err := suite.repo.Delete("not-exist")
	assert.NoError(suite.T(), err)
}

func (suite *UserRepoSuite) TestBeginTransactionError() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()

	_, err := suite.repo.BeginTransaction(context.Background())
	assert.Error(suite.T(), err)
}

func (suite *UserRepoSuite) TestBeginAndWithTransaction_Rollback() {
	tx, err := suite.repo.BeginTransaction(suite.T().Context())
	assert.NoError(suite.T(), err)

	txRepo := suite.repo.WithTransaction(tx)
	_, err = txRepo.Create("ivan", "hash", "ivan@example.com", entities.Developer, 1)
	assert.NoError(suite.T(), err)

	tx.Rollback()

	_, err = suite.repo.FindByName("ivan")
	assert.Error(suite.T(), err)
}
