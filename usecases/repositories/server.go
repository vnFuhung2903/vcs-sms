package repositories

import (
	"fmt"

	"github.com/vnFuhung2903/vcs-sms/entities"
	"gorm.io/gorm"
)

type ServerRepository interface {
	FindById(serverId string) (*entities.Server, error)
	Filter(filter *entities.ServerFilter, from int, to int, sortOpt entities.ServerSort) ([]*entities.Server, error)
	Create(serverId string, serverName string, ipv4 string) (*entities.Server, error)
	// Update(server *entities.Server) error
	Delete(serverId string) error
}

type serverRepository struct {
	Db *gorm.DB
}

func NewServerRepository(db *gorm.DB) ServerRepository {
	return &serverRepository{Db: db}
}

func (ur *serverRepository) FindById(serverId string) (*entities.Server, error) {
	var server entities.Server
	res := ur.Db.First(&server, entities.Server{ServerId: serverId})
	if res.Error != nil {
		return nil, res.Error
	}
	return &server, nil
}

func (ur *serverRepository) Filter(filter *entities.ServerFilter, from int, to int, sortOpt entities.ServerSort) ([]*entities.Server, error) {
	var servers []*entities.Server
	query := ur.Db.Model(&entities.Server{})

	if filter.ServerId != nil {
		query = query.Where("server_id = ?", *filter.ServerId)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.ServerName != nil {
		query = query.Where("server_name LIKE ?", *filter.ServerName)
	}
	if filter.Ipv4 != nil {
		query = query.Where("ipv4 = ?", *filter.Ipv4)
	}

	query = query.Limit(to - from + 1).Offset(from - 1).Find(&servers)
	query = query.Order(fmt.Sprintf("%s %s", sortOpt.Field, sortOpt.Sort))
	if query.Error != nil {
		return nil, query.Error
	}
	return servers, nil
}

func (ur *serverRepository) Create(serverId string, serverName string, ipv4 string) (*entities.Server, error) {
	newServer := &entities.Server{
		ServerId:   serverId,
		ServerName: serverName,
		Ipv4:       ipv4,
	}
	res := ur.Db.Create(newServer)
	if res.Error != nil {
		return nil, res.Error
	}
	return newServer, nil
}

// func (ur *serverRepository) Update(server *entities.Server, email string) error {
// 	res := ur.Db.Model(server).Update("email", email)
// 	return res.Error
// }

func (ur *serverRepository) Delete(serverId string) error {
	res := ur.Db.Delete(&entities.Server{ServerId: serverId})
	return res.Error
}
