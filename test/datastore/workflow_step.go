package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowStepCRUD(t *testing.T, workflowStepDAO dao.WorkflowStepDAO,
	org *config.Organization) {

	farmID := org.GetFarms()[0].GetID()

	drainStep := config.NewWorkflowStep()
	drainStep.SetWorkflowID(1)
	drainStep.SetDeviceID(1)
	drainStep.SetChannelID(0)
	drainStep.SetDuration(300) // seconds; 5 minutes
	drainStep.SetWait(300)     // seconds; 5 minutes

	err := workflowStepDAO.Save(0, drainStep)
	assert.Nil(t, err)

	persistedSteps, err := workflowStepDAO.GetByWorkflowID(farmID,
		drainStep.GetWorkflowID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedSteps))

	persistedSteps1 := persistedSteps[0]

	assert.Equal(t, uint64(1), persistedSteps1.GetID())
	assert.Equal(t, uint64(1), persistedSteps1.GetDeviceID())
}
