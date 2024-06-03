package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/datastore/dao"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowStepCRUD(t *testing.T, workflowStepDAO dao.WorkflowStepDAO,
	org *config.OrganizationStruct) {

	farm := org.GetFarms()[0]
	farmID := farm.ID

	drainStep := config.NewWorkflowStep()
	drainStep.SetWorkflowID(farm.Workflows[0].ID)
	drainStep.SetDeviceID(1)
	drainStep.SetChannelID(0)
	drainStep.SetDuration(300) // seconds; 5 minutes
	drainStep.SetWait(300)     // seconds; 5 minutes

	err := workflowStepDAO.Save(farmID, drainStep)
	assert.Nil(t, err)

	persistedSteps, err := workflowStepDAO.GetByWorkflowID(farmID,
		drainStep.GetWorkflowID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedSteps))

	persistedSteps1 := persistedSteps[0]

	assert.Greater(t, persistedSteps1.ID, uint64(0))
	assert.Equal(t, uint64(1), persistedSteps1.GetDeviceID())
}
