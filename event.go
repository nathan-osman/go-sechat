package sechat

import (
	"fmt"
)

const (
	EventMessagePosted           = 1
	EventMessageEdited           = 2
	EventUserJoined              = 3
	EventUserLeft                = 4
	EventRoomNameChanged         = 5
	EventMessageStarred          = 6
	EventDebugMessage            = 7
	EventUserMentioned           = 8
	EventMessageFlagged          = 9
	EventMessageDeleted          = 10
	EventFileAdded               = 11
	EventModeratorFlag           = 12
	EventUserSettingsChanged     = 13
	EventGlobalNotification      = 14
	EventAccessLevelChanged      = 15
	EventUserNotification        = 16
	EventInvitation              = 17
	EventMessageReply            = 18
	EventMessageMovedOut         = 19
	EventMessageMovedIn          = 20
	EventTimeBreak               = 21
	EventFeedTicker              = 22
	EventUserSuspended           = 29
	EventUserMerged              = 30
	EventUserNameOrAvatarChanged = 34
)

// Event represents an individual chat event received from the server. Note
// that not all event types use all members of this struct.
type Event struct {
	Content      string `json:"content"`
	EventType    int    `json:"event_type"`
	ID           int    `json:"id"`
	MessageEdits int    `json:"message_edits"`
	MessageID    int    `json:"message_id"`
	MessageStars int    `json:"message_stars"`
	Moved        bool   `json:"moved"`
	ParentID     int    `json:"parent_id"`
	RoomID       int    `json:"room_id"`
	RoomName     string `json:"room_name"`
	ShowParent   bool   `json:"show_parent"`
	TargetUserID int    `json:"target_user_id"`
	TimeStamp    int    `json:"time_stamp"`
	UserID       int    `json:"user_id"`
	UserName     string `json:"user_name"`
}

// IsMention determines whether the event mentions the user or not.
func (e *Event) IsMention() bool {
	return e.EventType == EventUserMentioned ||
		e.EventType == EventMessageReply
}

// FormatReply generates a reply to a message with the specified content.
func (e *Event) FormatReply(format string, a ...interface{}) string {
	return fmt.Sprintf(":%d ", e.MessageID) +
		fmt.Sprintf(format, a...)
}
