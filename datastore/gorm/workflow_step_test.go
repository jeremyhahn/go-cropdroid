package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowStepCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.Workflow{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStep{})

	workflowStepDAO := NewWorkflowStepDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, workflowStepDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)

	dstest.TestWorkflowStepCRUD(t, workflowStepDAO, org)
}
