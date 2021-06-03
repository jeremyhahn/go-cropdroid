// +build integrationpackage test

import (
	"testing"
	"time"

	"github.com/jeremyhahn/cropdroid/model"
	"github.com/jeremyhahn/cropdroid/service"
	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {

	ctx := NewUnitTestContext()
	mailer := mock(Mailer)

	notification := &model.Notification{
		Controller: "test",
		Type:       "T",
		Message:    "Test message",
		Timestamp:  time.Now()}

	service := service.NewNotificationService(ctx, mailer)
	service.Enqueue(notification)

	dequeued := <-service.Dequeue()

	assert.Equal(t, notification.GetController(), dequeued.GetController())
	assert.Equal(t, notification.GetType(), dequeued.GetType())
	assert.Equal(t, notification.GetMessage(), dequeued.GetMessage())
	assert.Equal(t, notification.GetTimestamp(), dequeued.GetTimestamp())
}
