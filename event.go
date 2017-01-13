package sechat

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

var mentionRegexp = regexp.MustCompile(
	`^@[\p{L}\d_]+`,
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

	// A few additional values are precomputed to simplify analysis later on
	IsMention   bool
	TextContent string
}

// precompute fills in the precomputed members.
func (e *Event) precompute() {
	e.IsMention = e.EventType == EventUserMentioned ||
		e.EventType == EventMessageReply
	if d, err := goquery.NewDocumentFromReader(
		strings.NewReader(e.Content),
	); err == nil {
		e.TextContent = strings.TrimSpace(
			mentionRegexp.ReplaceAllString(d.Text(), ""),
		)
	}
}
