package config

type Server struct {
	ID               uint64
	OrganizationRefs []uint64 `yaml:"organizationRefs" json:"organizationRefs" mapstructure:"organizations"`
	FarmRefs         []uint64 `yaml:"farmRefs" json:"farmRefs" mapstructure:"farms"`
}

func NewServer() *Server {
	return &Server{
		OrganizationRefs: make([]uint64, 0),
		FarmRefs:         make([]uint64, 0)}
}

func (server *Server) SetID(id uint64) {
	server.ID = id
}

func (server *Server) GetID() uint64 {
	return server.ID
}

func (server *Server) SetOrganizationRefs(refs []uint64) {
	server.OrganizationRefs = refs
}

func (server *Server) GetOrganizationRefs() []uint64 {
	return server.OrganizationRefs
}

func (server *Server) AddOrganizationRef(orgID uint64) {
	server.OrganizationRefs = append(server.OrganizationRefs, orgID)
}

func (server *Server) RemoveOrganizationRef(orgID uint64) {
	newRefs := make([]uint64, 0)
	for _, ref := range server.OrganizationRefs {
		if ref == orgID {
			continue
		}
		newRefs = append(newRefs, ref)
	}
	server.OrganizationRefs = newRefs
}

func (server *Server) SetFarmRefs(refs []uint64) {
	server.FarmRefs = refs
}

func (server *Server) GetFarmRefs() []uint64 {
	return server.FarmRefs
}

func (server *Server) AddFarmRef(farmID uint64) {
	server.FarmRefs = append(server.FarmRefs, farmID)
}

func (server *Server) HasFarmRef(farmID uint64) bool {
	for _, id := range server.FarmRefs {
		if id == farmID {
			return true
		}
	}
	return false
}

func (server *Server) RemoveFarmRef(farmID uint64) {
	newRefs := make([]uint64, 0)
	for _, ref := range server.FarmRefs {
		if ref == farmID {
			continue
		}
		newRefs = append(newRefs, ref)
	}
	server.FarmRefs = newRefs
}
