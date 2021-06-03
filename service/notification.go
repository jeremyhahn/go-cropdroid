package service

import (
	"errors"

	"github.com/jeremyhahn/cropdroid/common"
	logging "github.com/op/go-logging"
)

type NotificationServiceImpl struct {
	logger        *logging.Logger
	notifications chan common.Notification
	mailer        common.Mailer
	NotificationService
}

func NewNotificationService(logger *logging.Logger, mailer common.Mailer) NotificationService {
	return &NotificationServiceImpl{
		logger:        logger,
		notifications: make(chan common.Notification, common.BUFFERED_CHANNEL_SIZE),
		mailer:        mailer}
}

func (ns *NotificationServiceImpl) QueueSize() int {
	return len(ns.notifications)
}

func (ns *NotificationServiceImpl) Enqueue(notification common.Notification) error {
	if ns.mailer != nil {
		ns.mailer.Send(notification.GetTitle(), notification.GetType(), notification.GetMessage())
	}
	ns.logger.Debugf("[NotificationService.Enqueue] Enqueuing notification %v+", notification)
	if notification.GetPriority() == common.NOTIFICATION_PRIORITY_HIGH {
		if ns.mailer != nil {
			ns.mailer.Send(notification.GetTitle(), notification.GetType(), notification.GetMessage())
		}
	}
	select {
	case ns.notifications <- notification:
		ns.logger.Debugf("[NotificationService.Enqueue] Queue size: %d", len(ns.notifications))
	default:
		errmsg := "Notification channel full, discarding..."
		ns.logger.Error(errmsg)
		return errors.New(errmsg)
	}
	return nil
}

func (ns *NotificationServiceImpl) Dequeue() <-chan common.Notification {
	return ns.notifications
}
