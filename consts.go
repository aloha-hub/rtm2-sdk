package rtm2_sdk

// uri
const (
	UriMessageSubscribe   = 0
	UriMessageUnsubscribe = 1
	UriMessagePublish     = 2
	UriMessageEvent       = 3

	UriStreamJoin       = 4
	UriStreamLeave      = 5
	UriStreamJoinTopic  = 6
	UriStreamLeaveTopic = 7
	UriStreamPublish    = 8
	UriStreamEvent      = 9
	UriStreamTopicEvent = 10
	UriStreamSubTopic   = 11
	UriStreamUnsubTopic = 12

	UriStorageOpChannelMetaData       = 24
	UriStorageGetChannelMetaData      = 25
	UriStorageOpUserMetaData          = 26
	UriStorageGetUserMetaData         = 27
	UriStorageSubscribeUserMetaData   = 28
	UriStorageUnSubscribeUserMetaData = 29
	UriStorageChannelEvent            = 30
	UriStorageUserEvent               = 31

	UriPresenceWhereNow    = 36
	UriPresenceWhoNow      = 37
	UriPresenceSetState    = 38
	UriPresenceGetState    = 39
	UriPresenceRemoveState = 40
	UriPresenceEvent       = 41

	UriLogin              = 100
	UriLogout             = 101
	UriConnectStateChange = 102
	UriSetParam           = 103
	UriRenewToken         = 104

	UriLockAcquire = 105
	UriLockGet     = 106
	UriLockRelease = 107
	UriLockRemove  = 108
	UriLockRevoke  = 109
	UriLockSet     = 110
	UriLockEvent   = 111

	UriCommonRequest = 0xFFE
	UriCommonResp    = 0xFFF

	invalidUri = -1
	unknownErr = 999

	kParamSidecarEndpoint = "golang_sidecar_endpoint"
	kParamSidecarPort     = "golang_sidecar_port"
	kParamSidecarPath     = "golang_sidecar_path"

	DefaultSidecarPort = 7001
)

// channel type
const (
	Msg    = 0
	Stream = 1
)

// topic event type
const (
	TopicEventSnapshot = 0
	TopicEventJoin     = 1
	TopicEventLeave    = 2
)

func IsEvent(uri int32) bool {
	switch uri {
	case UriMessageEvent, UriStreamEvent, UriStreamTopicEvent, UriStorageChannelEvent, UriStorageUserEvent, UriConnectStateChange, UriPresenceEvent, UriLockEvent:
		return true
	}
	return false
}
