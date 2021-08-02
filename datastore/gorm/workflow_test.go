package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

	workflowDAO := NewWorkflowDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, workflowDAO)

	farmID := uint64(1)
	reservoirID := uint64(3)
	doserID := uint64(4)

	// Water change workflow
	drainStep := config.NewWorkflowStep()
	drainStep.SetDeviceID(reservoirID)
	drainStep.SetChannelID(0)  // CHANNEL_RESERVOIR_DRAIN
	drainStep.SetDuration(300) // seconds; 5 minutes
	drainStep.SetWait(300)     // seconds; 5 minutes

	fillStep := config.NewWorkflowStep()
	fillStep.SetDeviceID(reservoirID)
	fillStep.SetChannelID(6)  // CHANNEL_RESERVOIR_FAUCET
	fillStep.SetDuration(300) // seconds; 5 minutes
	fillStep.SetWait(300)     // seconds; 5 minutes

	phDownStep := config.NewWorkflowStep()
	phDownStep.SetDeviceID(doserID)
	phDownStep.SetChannelID(0) // CHANNEL_DOSER_PHDOWN
	phDownStep.SetDuration(60) // seconds
	phDownStep.SetWait(300)    // seconds; 5 minutes

	nutePart1Step := config.NewWorkflowStep()
	nutePart1Step.SetDeviceID(doserID)
	nutePart1Step.SetChannelID(4) // CHANNEL_DOSER_NUTE1
	nutePart1Step.SetDuration(30) // seconds
	nutePart1Step.SetWait(300)    // seconds; 5 minutes

	nutePart2Step := config.NewWorkflowStep()
	nutePart2Step.SetDeviceID(doserID)
	nutePart2Step.SetChannelID(5) // CHANNEL_DOSER_NUTE2
	nutePart2Step.SetDuration(30) // seconds
	nutePart2Step.SetWait(300)    // seconds; 5 minutes

	nutePart3Step := config.NewWorkflowStep()
	nutePart3Step.SetDeviceID(doserID)
	nutePart3Step.SetChannelID(6) // CHANNEL_DOSER_NUTE3
	nutePart3Step.SetDuration(30) // seconds
	nutePart3Step.SetWait(300)    // seconds; 5 minutes

	waterChangeWorkflow := config.NewWorkflow()
	waterChangeWorkflow.SetFarmID(farmID)
	waterChangeWorkflow.SetName("Automated Water Changes")
	waterChangeWorkflow.SetSteps([]config.WorkflowStep{
		*drainStep,
		*fillStep,
		*phDownStep,
		*nutePart1Step,
		*nutePart2Step,
		*nutePart3Step})

	err := workflowDAO.Create(waterChangeWorkflow)
	assert.Nil(t, err)

	persistedWorkflows, err := workflowDAO.GetByFarmID(farmID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedWorkflows))

	persistedWorkflow1 := persistedWorkflows[0]
	persistedSteps := persistedWorkflow1.GetSteps()

	assert.Equal(t, uint64(1), persistedWorkflow1.GetID())
	assert.Equal(t, waterChangeWorkflow.FarmID, persistedWorkflow1.GetFarmID())
	assert.Equal(t, len(waterChangeWorkflow.GetSteps()), len(persistedSteps))

	persistedStep1 := persistedSteps[0]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep1.WorkflowID)
	assert.Equal(t, drainStep.DeviceID, persistedStep1.GetDeviceID())
	assert.Equal(t, drainStep.ChannelID, persistedStep1.GetChannelID())
	assert.Equal(t, drainStep.DeviceID, persistedStep1.GetDeviceID())
	assert.Equal(t, drainStep.Duration, persistedStep1.GetDuration())
	assert.Equal(t, drainStep.Wait, persistedStep1.GetWait())

	persistedStep2 := persistedSteps[1]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep2.WorkflowID)
	assert.Equal(t, fillStep.DeviceID, persistedStep2.GetDeviceID())
	assert.Equal(t, fillStep.ChannelID, persistedStep2.GetChannelID())
	assert.Equal(t, fillStep.DeviceID, persistedStep2.GetDeviceID())
	assert.Equal(t, fillStep.Duration, persistedStep2.GetDuration())
	assert.Equal(t, fillStep.Wait, persistedStep2.GetWait())

	persistedStep3 := persistedSteps[2]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep3.WorkflowID)
	assert.Equal(t, phDownStep.DeviceID, persistedStep3.GetDeviceID())
	assert.Equal(t, phDownStep.ChannelID, persistedStep3.GetChannelID())
	assert.Equal(t, phDownStep.DeviceID, persistedStep3.GetDeviceID())
	assert.Equal(t, phDownStep.Duration, persistedStep3.GetDuration())
	assert.Equal(t, phDownStep.Wait, persistedStep3.GetWait())

	persistedStep4 := persistedSteps[3]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep4.WorkflowID)
	assert.Equal(t, nutePart1Step.DeviceID, persistedStep4.GetDeviceID())
	assert.Equal(t, nutePart1Step.ChannelID, persistedStep4.GetChannelID())
	assert.Equal(t, nutePart1Step.DeviceID, persistedStep4.GetDeviceID())
	assert.Equal(t, nutePart1Step.Duration, persistedStep4.GetDuration())
	assert.Equal(t, nutePart1Step.Wait, persistedStep4.GetWait())

	persistedStep5 := persistedSteps[4]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep5.WorkflowID)
	assert.Equal(t, nutePart2Step.DeviceID, persistedStep5.GetDeviceID())
	assert.Equal(t, nutePart2Step.ChannelID, persistedStep5.GetChannelID())
	assert.Equal(t, nutePart2Step.DeviceID, persistedStep5.GetDeviceID())
	assert.Equal(t, nutePart2Step.Duration, persistedStep5.GetDuration())
	assert.Equal(t, nutePart2Step.Wait, persistedStep5.GetWait())

	persistedStep6 := persistedSteps[5]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep6.WorkflowID)
	assert.Equal(t, nutePart3Step.DeviceID, persistedStep6.GetDeviceID())
	assert.Equal(t, nutePart3Step.ChannelID, persistedStep6.GetChannelID())
	assert.Equal(t, nutePart3Step.DeviceID, persistedStep6.GetDeviceID())
	assert.Equal(t, nutePart3Step.Duration, persistedStep6.GetDuration())
	assert.Equal(t, nutePart3Step.Wait, persistedStep6.GetWait())

	currentTest.Cleanup()
}
