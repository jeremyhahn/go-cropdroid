package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowStepCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	//currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

	workflowStepDAO := NewWorkflowStepDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, workflowStepDAO)

	// Water change workflow
	drainStep := config.NewWorkflowStep()
	drainStep.SetWorkflowID(1)
	drainStep.SetDeviceID(1)
	drainStep.SetChannelID(0)  // CHANNEL_RESERVOIR_DRAIN
	drainStep.SetDuration(300) // seconds; 5 minutes
	drainStep.SetWait(300)     // seconds; 5 minutes

	err := workflowStepDAO.Create(drainStep)
	assert.Nil(t, err)

	persistedSteps, err := workflowStepDAO.GetByWorkflowID(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedSteps))

	persistedSteps1 := persistedSteps[0]

	assert.Equal(t, uint64(1), persistedSteps1.GetID())
	assert.Equal(t, uint64(1), persistedSteps1.GetDeviceID())

	currentTest.Cleanup()
}
