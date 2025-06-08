package repositories

import (
	"context"
	"fmt"

	"github.com/vnFuhung2903/vcs-sms/entities"
	"gorm.io/gorm"
)

type IServerRepository interface {
	FindById(serverId string) (*entities.Server, error)
	FindByName(serverName string) (*entities.Server, error)
	View(filter entities.ServerFilter, from int, limit int, sortOpt entities.ServerSort) ([]*entities.Server, int64, error)
	Create(serverId string, serverName string, status entities.ServerStatus, ipv4 string) (*entities.Server, error)
	Update(server *entities.Server, updateData map[string]interface{}) error
	Delete(serverId string) error
	BeginTransaction(ctx context.Context) (*gorm.DB, error)
	WithTransaction(tx *gorm.DB) IServerRepository
}

type serverRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) IServerRepository {
	return &serverRepository{db: db}
}

func (r *serverRepository) FindById(serverId string) (*entities.Server, error) {
	var server entities.Server
	res := r.db.First(&server, entities.Server{ServerId: serverId})
	if res.Error != nil {
		return nil, res.Error
	}
	return &server, nil
}

func (r *serverRepository) FindByName(serverName string) (*entities.Server, error) {
	var server entities.Server
	res := r.db.First(&server, entities.Server{ServerName: serverName})
	if res.Error != nil {
		return nil, res.Error
	}
	return &server, nil
}

func (r *serverRepository) View(filter entities.ServerFilter, from int, limit int, sort entities.ServerSort) (servers []*entities.Server, total int64, err error) {
	query := r.db.Model(entities.Server{})

	if filter.ServerId != "" {
		query = query.Where("server_id = ?", filter.ServerId)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ServerName != "" {
		query = query.Where("server_name LIKE ?", "%"+filter.ServerName+"%")
	}
	if filter.Ipv4 != "" {
		query = query.Where("ipv4 = ?", filter.Ipv4)
	}

	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order(fmt.Sprintf("%s %s", sort.Field, sort.Sort))
	if err = query.Limit(limit).Offset(from - 1).Find(&servers).Error; err != nil {
		return nil, 0, err
	}
	return
}

func (r *serverRepository) Create(serverId string, serverName string, status entities.ServerStatus, ipv4 string) (*entities.Server, error) {
	newServer := &entities.Server{
		ServerId:   serverId,
		Status:     status,
		ServerName: serverName,
		Ipv4:       ipv4,
	}
	res := r.db.Create(newServer)
	if res.Error != nil {
		return nil, res.Error
	}
	return newServer, nil
}

func (r *serverRepository) Update(server *entities.Server, updateData map[string]interface{}) error {
	res := r.db.Model(server).Updates(updateData)
	return res.Error
}

func (r *serverRepository) Delete(serverId string) error {
	res := r.db.Delete(&entities.Server{ServerId: serverId})
	return res.Error
}

func (r *serverRepository) BeginTransaction(ctx context.Context) (*gorm.DB, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

func (r *serverRepository) WithTransaction(tx *gorm.DB) IServerRepository {
	return &serverRepository{db: tx}
}
