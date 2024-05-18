package raft

import (
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
)

type RaftUserDAO interface {
	RaftDAO[*config.User]
	dao.UserDAO
}

type RaftAlgorithmDAO interface {
	RaftDAO[*config.Algorithm]
	dao.AlgorithmDAO
}
