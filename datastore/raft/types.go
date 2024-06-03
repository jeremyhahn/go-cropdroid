package raft

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
)

type RaftUserDAO interface {
	RaftDAO[*config.UserStruct]
	dao.UserDAO
}

type RaftAlgorithmDAO interface {
	RaftDAO[*config.AlgorithmStruct]
	dao.AlgorithmDAO
}
