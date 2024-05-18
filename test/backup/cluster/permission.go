//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"fmt"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	logging "github.com/op/go-logging"
)

type RaftPermissionDAO struct {
	logger          *logging.Logger
	organizationDAO dao.OrganizationDAO
	farmDAO         dao.FarmDAO
	userDAO         dao.UserDAO
	dao.PermissionDAO
}

func NewRaftPermissionDAO(logger *logging.Logger,
	organizationDAO dao.OrganizationDAO,
	farmDAO dao.FarmDAO, userDAO dao.UserDAO) dao.PermissionDAO {
	return &RaftPermissionDAO{
		logger:          logger,
		organizationDAO: organizationDAO,
		farmDAO:         farmDAO,
		userDAO:         userDAO}
}

// Returns all of the organizations for the specified user
func (permissionDAO *RaftPermissionDAO) GetOrganizations(userID uint64,
	CONSISTENCY_LEVEL int) ([]*config.Organization, error) {

	user, err := permissionDAO.userDAO.Get(userID, common.CONSISTENCY_LOCAL)
	if err != nil {
		return nil, err
	}
	orgIDs := user.GetOrganizationRefs()
	orgs := make([]*config.Organization, len(orgIDs))
	for i, orgID := range orgIDs {
		org, err := permissionDAO.organizationDAO.Get(orgID, CONSISTENCY_LEVEL)
		if err != nil {
			return nil, err
		}
		orgs[i] = org
	}
	return orgs, nil
}

// Returns all users belonging to the specified organization id
func (permissionDAO *RaftPermissionDAO) GetUsers(orgID uint64,
	CONSISTENCY_LEVEL int) ([]*config.User, error) {

	org, err := permissionDAO.organizationDAO.Get(orgID, CONSISTENCY_LEVEL)
	if err != nil {
		return nil, err
	}
	return org.GetUsers(), nil
}

func (permissionDAO *RaftPermissionDAO) GetFarms(orgID uint64,
	CONSISTENCY_LEVEL int) ([]*config.Farm, error) {

	org, err := permissionDAO.organizationDAO.Get(orgID, CONSISTENCY_LEVEL)
	if err != nil {
		return nil, err
	}
	return org.GetFarms(), nil
}

func (permissionDAO *RaftPermissionDAO) Save(permission *config.Permission) error {
	permissionDAO.logger.Debugf("Saving permission: %+v", permission)

	// perm, err := json.Marshal(*permission.(*config.Permission))
	// if err != nil {
	// 	permissionDAO.logger.Errorf("Error: %s", err)
	// 	return err
	// }
	// proposal, err := statemachine.CreateProposal(
	// 	statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	// if err != nil {
	// 	permissionDAO.logger.Errorf("Error: %s", err)
	// 	return err
	// }
	// if err := permissionDAO.raft.SyncPropose(permissionDAO.clusterID, proposal); err != nil {
	// 	permissionDAO.logger.Errorf("Error: %s", err)
	// 	return err
	// }

	user, err := permissionDAO.userDAO.Get(permission.GetUserID(), common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}

	// Organization membership
	belongsToOrg := false
	for _, orgID := range user.GetOrganizationRefs() {
		if orgID == permission.GetOrgID() {
			belongsToOrg = true
			break
		}
	}
	if !belongsToOrg && permission.GetOrgID() > 0 {
		user.AddOrganizationRef(permission.GetOrgID())
		// Add the user to the organization
		org, err := permissionDAO.organizationDAO.Get(permission.GetOrgID(), common.CONSISTENCY_LOCAL)
		if err != nil {
			return err
		}
		org.AddUser(user)
		if err := permissionDAO.organizationDAO.Save(org); err != nil {
			return err
		}
	}

	// Farm membership
	belongsToFarm := false
	for _, farmID := range user.GetFarmRefs() {
		if farmID == permission.GetFarmID() {
			belongsToFarm = true
			break
		}
	}
	if !belongsToFarm {
		user.AddFarmRef(permission.GetFarmID())
		// Add the user to the farm
		farm, err := permissionDAO.farmDAO.Get(permission.GetFarmID(), common.CONSISTENCY_LOCAL)
		if err != nil {
			return err
		}
		farm.AddUser(user)
		if err := permissionDAO.farmDAO.Save(farm); err != nil {
			return err
		}
	}

	// Role membership
	// belongsToRole := false
	// for _, role := range user.GetRoles() {
	// 	if role.GetID() == permission.GetRoleID() {
	// 		belongsToRole = true
	// 		break
	// 	}
	// }
	// if !belongsToRole {
	// 	role, err := permissionDAO.roleDAO.Get(permission.GetRoleID(), common.CONSISTENCY_LOCAL)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	user.AddRole(role)
	// }

	// Save to database
	if !belongsToOrg || !belongsToFarm /*|| !belongsToRole */ {
		if err := permissionDAO.userDAO.Save(user); err != nil {
			return err
		}
	}

	return nil
}

func (permissionDAO *RaftPermissionDAO) Update(permission *config.Permission) error {
	permissionDAO.logger.Debugf(fmt.Sprintf("Updating permission record: %+v", permission))
	// perm, err := json.Marshal(*permission.(*config.Permission))
	// if err != nil {
	// 	permissionDAO.logger.Errorf("Error: %s", err)
	// 	return err
	// }
	// proposal, err := statemachine.CreateProposal(
	// 	statemachine.QUERY_TYPE_UPDATE, perm).Serialize()
	// if err != nil {
	// 	permissionDAO.logger.Errorf("Error: %s", err)
	// 	return err
	// }
	// if err := permissionDAO.raft.SyncPropose(permissionDAO.clusterID, proposal); err != nil {
	// 	permissionDAO.logger.Errorf("Error: %s", err)
	// 	return err
	// }
	return permissionDAO.Save(permission)
}

func (permissionDAO *RaftPermissionDAO) Delete(permission *config.Permission) error {
	permissionDAO.logger.Debugf(fmt.Sprintf("Deleting permission record: %+v", permission))
	// perm, err := json.Marshal(*permission.(*config.Permission))
	// if err != nil {
	// 	permissionDAO.logger.Errorf("Error: %s", err)
	// 	return err
	// }
	// proposal, err := statemachine.CreateProposal(
	// 	statemachine.QUERY_TYPE_DELETE, perm).Serialize()
	// if err != nil {
	// 	permissionDAO.logger.Errorf("Error: %s", err)
	// 	return err
	// }
	// if err := permissionDAO.raft.SyncPropose(permissionDAO.clusterID, proposal); err != nil {
	// 	permissionDAO.logger.Errorf("Error: %s", err)
	// 	return err
	// }
	// return nil

	isUserModified := false

	user, err := permissionDAO.userDAO.Get(permission.GetUserID(), common.CONSISTENCY_LOCAL)
	if err != nil {
		return err
	}

	// Organization membership
	newOrgMembership := make([]uint64, 0)
	for _, orgID := range user.GetOrganizationRefs() {
		if orgID == permission.GetOrgID() {
			continue
		}
		newOrgMembership = append(newOrgMembership, orgID)
	}
	if len(newOrgMembership) != len(user.GetOrganizationRefs()) {
		user.SetOrganizationRefs(newOrgMembership)
		isUserModified = true

		// Remove the user from the organization
		org, err := permissionDAO.organizationDAO.Get(permission.GetOrgID(), common.CONSISTENCY_LOCAL)
		if err != nil {
			return err
		}
		org.RemoveUser(user)
		if err := permissionDAO.organizationDAO.Save(org); err != nil {
			return err
		}
	}

	// Farm membership
	newFarmMembership := make([]uint64, 0)
	for _, farmID := range user.GetFarmRefs() {
		if farmID == permission.GetFarmID() {
			continue
		}
		newFarmMembership = append(newFarmMembership, farmID)
	}
	if len(newFarmMembership) != len(user.GetFarmRefs()) {
		user.SetFarmRefs(newFarmMembership)
		isUserModified = true

		// Remove the user from the farm
		farm, err := permissionDAO.farmDAO.Get(permission.GetOrgID(), common.CONSISTENCY_LOCAL)
		if err != nil {
			return err
		}
		farm.RemoveUser(user)
		if err := permissionDAO.farmDAO.Save(farm); err != nil {
			return err
		}
	}

	// Role membership
	// newRoleMembership := make([]config.RoleConfig, 0)
	// for _, role := range user.GetRoles() {
	// 	if role.GetID() == permission.GetRoleID() {
	// 		continue
	// 	}
	// 	newRoleMembership = append(newRoleMembership, role)
	// }
	// if len(newRoleMembership) != len(user.GetRoles()) {
	// 	role, err := permissionDAO.roleDAO.Get(permission.GetRoleID(), common.CONSISTENCY_LOCAL)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	user.RemoveRole(role)
	// 	isUserModified = true
	// }

	// Save to database
	if isUserModified {
		if err := permissionDAO.userDAO.Save(user); err != nil {
			return err
		}
	}

	return nil
}
