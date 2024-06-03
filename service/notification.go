package service

import (
	"errors"

	"github.com/jeremyhahn/go-cropdroid/common"
	"github.com/jeremyhahn/go-cropdroid/model"
	logging "github.com/op/go-logging"
)

type NotificationServicer interface {
	Enqueue(notification model.Notification) error
	Dequeue() <-chan model.Notification
	QueueSize() int
}

type NotificationService struct {
	logger        *logging.Logger
	notifications chan model.Notification
	mailer        common.Mailer
	NotificationServicer
}

func NewNotificationService(
	logger *logging.Logger,
	mailer common.Mailer) NotificationServicer {

	return &NotificationService{
		logger:        logger,
		notifications: make(chan model.Notification, common.BUFFERED_CHANNEL_SIZE),
		mailer:        mailer}
}

func (ns *NotificationService) QueueSize() int {
	return len(ns.notifications)
}

func (ns *NotificationService) Enqueue(notification model.Notification) error {
	if ns.mailer != nil {
		ns.mailer.Send(notification.GetType(), notification.GetMessage())
	}
	ns.logger.Debugf("Enqueuing notification %v+", notification)
	if notification.GetPriority() == common.NOTIFICATION_PRIORITY_HIGH {
		if ns.mailer != nil {
			ns.mailer.Send(notification.GetType(), notification.GetMessage())
		}
	}
	select {
	case ns.notifications <- notification:
		ns.logger.Debugf("Queue size: %d", len(ns.notifications))
	default:
		errmsg := "Notification channel full, discarding..."
		ns.logger.Error(errmsg)
		return errors.New(errmsg)
	}
	return nil
}

func (ns *NotificationService) Dequeue() <-chan model.Notification {
	return ns.notifications
}
