package datastore

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/config"
	"github.com/jeremyhahn/go-cropdroid/config/dao"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowCRUD(t *testing.T, workflowDAO dao.WorkflowDAO,
	org *config.Organization) {

	farm := org.GetFarms()[0]
	farmID := farm.GetID()
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
	waterChangeWorkflow.SetSteps([]*config.WorkflowStep{
		drainStep,
		fillStep,
		phDownStep,
		nutePart1Step,
		nutePart2Step,
		nutePart3Step})

	err := workflowDAO.Save(waterChangeWorkflow)
	assert.Nil(t, err)

	persistedWorkflows, err := workflowDAO.GetByFarmID(
		waterChangeWorkflow.GetFarmID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(persistedWorkflows))

	persistedWorkflow1 := persistedWorkflows[0]
	persistedSteps := persistedWorkflow1.GetSteps()

	//assert.Equal(t, uint64(1), persistedWorkflow1.GetID())
	assert.Equal(t, waterChangeWorkflow.GetFarmID(), persistedWorkflow1.GetFarmID())
	assert.Equal(t, len(waterChangeWorkflow.GetSteps()), len(persistedSteps))

	persistedStep1 := persistedSteps[0]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep1.GetWorkflowID())
	assert.Equal(t, drainStep.GetDeviceID(), persistedStep1.GetDeviceID())
	assert.Equal(t, drainStep.GetChannelID(), persistedStep1.GetChannelID())
	assert.Equal(t, drainStep.GetDeviceID(), persistedStep1.GetDeviceID())
	assert.Equal(t, drainStep.GetDuration(), persistedStep1.GetDuration())
	assert.Equal(t, drainStep.GetWait(), persistedStep1.GetWait())

	persistedStep2 := persistedSteps[1]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep2.GetWorkflowID())
	assert.Equal(t, fillStep.GetDeviceID(), persistedStep2.GetDeviceID())
	assert.Equal(t, fillStep.GetChannelID(), persistedStep2.GetChannelID())
	assert.Equal(t, fillStep.GetDeviceID(), persistedStep2.GetDeviceID())
	assert.Equal(t, fillStep.GetDuration(), persistedStep2.GetDuration())
	assert.Equal(t, fillStep.GetWait(), persistedStep2.GetWait())

	persistedStep3 := persistedSteps[2]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep3.GetWorkflowID())
	assert.Equal(t, phDownStep.GetDeviceID(), persistedStep3.GetDeviceID())
	assert.Equal(t, phDownStep.GetChannelID(), persistedStep3.GetChannelID())
	assert.Equal(t, phDownStep.GetDeviceID(), persistedStep3.GetDeviceID())
	assert.Equal(t, phDownStep.GetDuration(), persistedStep3.GetDuration())
	assert.Equal(t, phDownStep.GetWait(), persistedStep3.GetWait())

	persistedStep4 := persistedSteps[3]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep4.GetWorkflowID())
	assert.Equal(t, nutePart1Step.GetDeviceID(), persistedStep4.GetDeviceID())
	assert.Equal(t, nutePart1Step.GetChannelID(), persistedStep4.GetChannelID())
	assert.Equal(t, nutePart1Step.GetDeviceID(), persistedStep4.GetDeviceID())
	assert.Equal(t, nutePart1Step.GetDuration(), persistedStep4.GetDuration())
	assert.Equal(t, nutePart1Step.GetWait(), persistedStep4.GetWait())

	persistedStep5 := persistedSteps[4]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep5.GetWorkflowID())
	assert.Equal(t, nutePart2Step.GetDeviceID(), persistedStep5.GetDeviceID())
	assert.Equal(t, nutePart2Step.GetChannelID(), persistedStep5.GetChannelID())
	assert.Equal(t, nutePart2Step.GetDeviceID(), persistedStep5.GetDeviceID())
	assert.Equal(t, nutePart2Step.GetDuration(), persistedStep5.GetDuration())
	assert.Equal(t, nutePart2Step.GetWait(), persistedStep5.GetWait())

	persistedStep6 := persistedSteps[5]
	assert.Equal(t, persistedWorkflow1.GetID(), persistedStep6.GetWorkflowID())
	assert.Equal(t, nutePart3Step.GetDeviceID(), persistedStep6.GetDeviceID())
	assert.Equal(t, nutePart3Step.GetChannelID(), persistedStep6.GetChannelID())
	assert.Equal(t, nutePart3Step.GetDeviceID(), persistedStep6.GetDeviceID())
	assert.Equal(t, nutePart3Step.GetDuration(), persistedStep6.GetDuration())
	assert.Equal(t, nutePart3Step.GetWait(), persistedStep6.GetWait())

	err = workflowDAO.Delete(persistedWorkflow1)
	assert.Nil(t, err)

	persistedWorkflows, err = workflowDAO.GetByFarmID(
		waterChangeWorkflow.GetFarmID(), common.CONSISTENCY_LOCAL)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(persistedWorkflows))
}
