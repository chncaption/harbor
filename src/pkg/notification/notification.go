package notification

import (
	"container/list"
	"context"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification/hook"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	n_event "github.com/goharbor/harbor/src/pkg/notifier/event"
	notifier_model "github.com/goharbor/harbor/src/pkg/notifier/model"
)

type (
	// EventType is the type of event
	EventType string
	// NotifyType is the type of notify
	NotifyType string
)

func (e EventType) String() string {
	return string(e)
}

func (n NotifyType) String() string {
	return string(n)
}

var (
	// PolicyMgr is a global notification policy manager
	PolicyMgr policy.Manager

	// HookManager is a hook manager
	HookManager hook.Manager

	// supportedEventTypes is a slice to store supported event type, eg. pushImage, pullImage etc
	supportedEventTypes []EventType

	// supportedNotifyTypes is a slice to store notification type, eg. HTTP, Email etc
	supportedNotifyTypes []NotifyType
)

// Init ...
func Init() {
	// init notification policy manager
	PolicyMgr = policy.Mgr
	// init hook manager
	HookManager = hook.NewHookManager()

	initSupportedNotifyType()

	log.Info("notification initialization completed")
}

func initSupportedNotifyType() {
	supportedEventTypes = make([]EventType, 0)
	supportedNotifyTypes = make([]NotifyType, 0)

	eventTypes := []string{
		event.TopicPushArtifact,
		event.TopicPullArtifact,
		event.TopicDeleteArtifact,
		event.TopicQuotaExceed,
		event.TopicQuotaWarning,
		event.TopicScanningFailed,
		event.TopicScanningStopped,
		event.TopicScanningCompleted,
		event.TopicReplication,
		event.TopicTagRetention,
	}
	for _, eventType := range eventTypes {
		supportedEventTypes = append(supportedEventTypes, EventType(eventType))
	}

	notifyTypes := []string{notifier_model.NotifyTypeHTTP, notifier_model.NotifyTypeSlack}
	for _, notifyType := range notifyTypes {
		supportedNotifyTypes = append(supportedNotifyTypes, NotifyType(notifyType))
	}
}

type eventKey struct{}

// EventCtx ...
type EventCtx struct {
	Events     *list.List
	MustNotify bool
}

// NewEventCtx returns instance of EventCtx
func NewEventCtx() *EventCtx {
	return &EventCtx{
		Events:     list.New(),
		MustNotify: false,
	}
}

// NewContext returns new context with event
func NewContext(ctx context.Context, ec *EventCtx) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, eventKey{}, ec)
}

// AddEvent add events into request context, the event will be sent by the notification middleware eventually.
func AddEvent(ctx context.Context, m n_event.Metadata, notify ...bool) {
	if m == nil {
		return
	}

	e, ok := ctx.Value(eventKey{}).(*EventCtx)
	if !ok {
		log.Debug("request has not event list, cannot add event into context")
		return
	}
	if len(notify) != 0 {
		e.MustNotify = notify[0]
	}
	e.Events.PushBack(m)
}

func GetSupportedEventTypes() []EventType {
	return supportedEventTypes
}

func GetSupportedNotifyTypes() []NotifyType {
	return supportedNotifyTypes
}
