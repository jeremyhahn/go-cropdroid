//go:build cluster && pebble
// +build cluster,pebble

package cluster

import (
	"fmt"
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/util"
	"github.com/stretchr/testify/assert"
	//dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestWorkflowStepCRUD(t *testing.T) {

	serverDAO := NewRaftServerDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), ClusterID)

	userDAO := NewRaftUserDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), UserClusterID)

	farmDAO := NewRaftFarmConfigDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), serverDAO, userDAO)

	workflowStepDAO := NewRaftWorkflowStepDAO(Cluster.app.Logger,
		Cluster.GetRaftNode1(), farmDAO)

	org := createRaftTestOrganization(t, Cluster, ClusterID,
		serverDAO, userDAO, farmDAO)

	farm1 := org.GetFarms()[0]
	farmID := farm1.GetID()

	// TODO: ConfigRefactor
	workflowName := "Test Workflow"
	workflowKey := fmt.Sprintf("%d-%s", farmID, workflowName)
	workflowID := Cluster.app.IdGenerator.NewStringID(workflowKey)
	// end TODO

	testWorkflow := config.NewWorkflow()
	testWorkflow.SetID(workflowID) // todo: remove
	testWorkflow.SetFarmID(farmID)
	testWorkflow.SetName(workflowName)

	farm1.SetWorkflow(testWorkflow)

	err := farmDAO.Save(farm1)
	assert.Nil(t, err)

	// Add a step
	testStep := config.NewWorkflowStep()
	testStep.SetWorkflowID(workflowID) // todo: remove
	testStep.SetDeviceID(1)
	testStep.SetChannelID(0)
	testStep.SetDuration(300) // seconds; 5 minutes
	testStep.SetWait(300)     // seconds; 5 minutes

	testWorkflow.SetSteps([]*config.WorkflowStep{testStep})

	err = workflowStepDAO.Save(farmID, testStep)
	assert.Nil(t, err)

	persistedSteps, err := workflowStepDAO.GetByWorkflowID(farmID,
		workflowID, common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedSteps))

	persistedSteps1 := persistedSteps[0]

	key := fmt.Sprintf("%d-%d-%d-%d-%d", persistedSteps1.GetWorkflowID(),
		persistedSteps1.GetDeviceID(), persistedSteps1.GetChannelID(),
		persistedSteps1.GetDuration(), persistedSteps1.GetState())
	idGenerator := util.NewIdGenerator(common.DATASTORE_TYPE_64BIT)
	id := idGenerator.NewStringID(key)

	assert.Equal(t, id, persistedSteps1.GetID())
	assert.Equal(t, uint64(1), persistedSteps1.GetDeviceID())

	//dstest.TestWorkflowStepCRUD(t, workflowStepDAO, org)
}
