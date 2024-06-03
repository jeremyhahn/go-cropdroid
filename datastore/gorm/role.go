package gorm

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/jeremyhahn/go-cropdroid/datastore/raft/query"
	logging "github.com/op/go-logging"
	"gorm.io/gorm"
)

type GormRoleDAO struct {
	logger *logging.Logger
	db     *gorm.DB
	dao.RoleDAO
}

func NewRoleDAO(logger *logging.Logger, db *gorm.DB) dao.RoleDAO {
	return &GormRoleDAO{logger: logger, db: db}
}

func (dao *GormRoleDAO) Save(role *config.RoleStruct) error {
	return dao.db.Save(&role).Error
}

func (dao *GormRoleDAO) Delete(role *config.RoleStruct) error {
	return dao.db.Delete(role).Error
}

// This method is only here for sake of completeness for the interface. This method
// is only used by the Raft datastore.
func (dao *GormRoleDAO) Get(roleID uint64, CONSISTENCY_LEVEL int) (*config.RoleStruct, error) {
	dao.logger.Debugf("Getting role %s", roleID)
	var role *config.RoleStruct
	if err := dao.db.
		First(&role, roleID).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			dao.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		dao.logger.Error(err)
		return nil, err
	}
	return role, nil
}

func (dao *GormRoleDAO) GetByName(name string, CONSISTENCY_LEVEL int) (*config.RoleStruct, error) {
	dao.logger.Debugf("Getting role %s", name)
	var role config.RoleStruct
	if err := dao.db.
		Table("roles").
		First(&role, "name = ?", name).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			dao.logger.Warning(err)
			return nil, datastore.ErrRecordNotFound
		}
		dao.logger.Error(err)
		return nil, err
	}
	return &role, nil
}

func (roleDAO *GormRoleDAO) GetPage(pageQuery query.PageQuery, CONSISTENCY_LEVEL int) (dao.PageResult[*config.RoleStruct], error) {
	roleDAO.logger.Debugf("Fetching role page: %v+", pageQuery)
	pageResult := dao.PageResult[*config.RoleStruct]{
		Page:     pageQuery.Page,
		PageSize: pageQuery.PageSize}
	page := pageQuery.Page
	if page < 1 {
		page = 1
	}
	var offset = (page - 1) * pageQuery.PageSize
	var roles []*config.RoleStruct
	if err := roleDAO.db.
		Offset(offset).
		Limit(pageQuery.PageSize + 1). // peek one record to set HasMore flag
		Find(&roles).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			roleDAO.logger.Warning(err)
			return pageResult, datastore.ErrRecordNotFound
		}
		roleDAO.logger.Error(err)
		return pageResult, err
	}
	// If the peek record was returned, set the HasMore flag and remove the +1 record
	if len(roles) == pageQuery.PageSize+1 {
		pageResult.HasMore = true
		roles = roles[:len(roles)-1]
	}
	pageResult.Entities = roles
	return pageResult, nil
}

func (roleDAO *GormRoleDAO) ForEachPage(pageQuery query.PageQuery,
	pagerProcFunc query.PagerProcFunc[*config.RoleStruct], CONSISTENCY_LEVEL int) error {

	pageResult, err := roleDAO.GetPage(pageQuery, CONSISTENCY_LEVEL)
	if err != nil {
		roleDAO.logger.Error(err)
		return nil
	}
	if err = pagerProcFunc(pageResult.Entities); err != nil {
		roleDAO.logger.Error(err)
		return err
	}
	if pageResult.HasMore {
		nextPageQuery := query.PageQuery{
			Page:      pageQuery.Page + 1,
			PageSize:  pageQuery.PageSize,
			SortOrder: pageQuery.SortOrder}
		return roleDAO.ForEachPage(nextPageQuery, pagerProcFunc, CONSISTENCY_LEVEL)
	}
	return nil
}
