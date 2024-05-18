//go:build cluster && pebble
// +build cluster,pebble

package raft

import (
	"testing"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestWorkflowStepCRUD(t *testing.T) {

	raftNode1 := IntegrationTestCluster.GetRaftNode1()
	org, _, farmDAO, _ := createRaftTestOrganization(t, IntegrationTestCluster, ClusterID)

	workflowStepDAO := NewRaftWorkflowStepDAO(
		IntegrationTestCluster.app.Logger,
		raftNode1,
		farmDAO)

	//armID := org.Farms[0].ID

	// testWorkflow := config.NewWorkflow()
	// testWorkflow.SetFarmID(farmID)
	// testWorkflow.SetName("Test Workflow")

	// org.Farms[0].SetWorkflow(testWorkflow)

	// err := farmDAO.Save(org.Farms[0])
	// assert.Nil(t, err)

	// // Add a step
	// testStep := config.NewWorkflowStep()
	// testStep.SetWorkflowID(testWorkflow.ID) // todo: remove
	// testStep.SetDeviceID(1)
	// testStep.SetChannelID(0)
	// testStep.SetDuration(300) // seconds; 5 minutes
	// testStep.SetWait(300)     // seconds; 5 minutes

	// testWorkflow.SetSteps([]*config.WorkflowStep{testStep})

	// err = workflowStepDAO.Save(farmID, testStep)
	// assert.Nil(t, err)

	// persistedSteps, err := workflowStepDAO.GetByWorkflowID(farmID,
	// 	testWorkflow.ID, common.CONSISTENCY_LOCAL)
	// assert.Nil(t, err)
	// assert.Equal(t, 1, len(persistedSteps))

	// persistedSteps1 := persistedSteps[0]
	// assert.Greater(t, persistedSteps1.ID, uint64(0))
	// assert.Equal(t, uint64(1), persistedSteps1.GetDeviceID())

	dstest.TestWorkflowStepCRUD(t, workflowStepDAO, org)
}
