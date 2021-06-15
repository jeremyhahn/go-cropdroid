package test

import (
	"testing"
	"time"

	"github.com/jeremyhahn/go-cropdroid/model"
	"github.com/jeremyhahn/go-cropdroid/service"
	"github.com/jeremyhahn/go-cropdroid/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {

	app, session := NewUnitTestSession()

	mailer := mocks.NewMockMailer(session)

	notification := &model.Notification{
		Device: "test",
		Type:       "T",
		Message:    "Test message",
		Timestamp:  time.Now()}

	service := service.NewNotificationService(app.Logger, mailer)
	service.Enqueue(notification)

	dequeued := <-service.Dequeue()

	assert.Equal(t, notification.GetDevice(), dequeued.GetDevice())
	assert.Equal(t, notification.GetType(), dequeued.GetType())
	assert.Equal(t, notification.GetMessage(), dequeued.GetMessage())
	assert.Equal(t, notification.GetTimestamp(), dequeued.GetTimestamp())
}
