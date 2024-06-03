package gorm

import (
	"testing"

	"github.com/jeremyhahn/go-cropdroid/config"

	"github.com/stretchr/testify/assert"

	dstest "github.com/jeremyhahn/go-cropdroid/test/datastore"
)

func TestWorkflowCRUD(t *testing.T) {

	currentTest := NewIntegrationTest()
	defer currentTest.Cleanup()

	currentTest.gorm.AutoMigrate(&config.WorkflowStruct{})
	currentTest.gorm.AutoMigrate(&config.WorkflowStepStruct{})

	workflowDAO := NewWorkflowDAO(currentTest.logger, currentTest.gorm)
	assert.NotNil(t, workflowDAO)

	org := dstest.CreateTestOrganization(currentTest.idGenerator)
	dstest.TestWorkflowCRUD(t, workflowDAO, org)
}
