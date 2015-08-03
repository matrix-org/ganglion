package types

import (
	"encoding/json"
	"strconv"
	"time"
)

type EventType string
type StateKey string

const (
	EventTypeCreate      EventType = "m.room.create"
	EventTypeName                  = "m.room.name"
	EventTypeTopic                 = "m.room.topic"
	EventTypeAliases               = "m.room.aliases"
	EventTypeJoinRules             = "m.room.join_rules"
	EventTypeMembership            = "m.room.member"
	EventTypePowerLevels           = "m.room.power_levels"
)

type TypedContent interface {
	EventType() EventType
}

type Content interface{}

type Timestamp struct {
	time.Time
}

type Event struct {
	EventId   EventId   `json:"event_id"`
	RoomId    RoomId    `json:"room_id"`
	UserId    UserId    `json:"user_id"`
	EventType EventType `json:"type"`
	Timestamp Timestamp `json:"origin_server_ts"`
	Content   Content   `json:"content"`
}

type OldState State

type State struct {
	Event
	StateKey StateKey  `json:"state_key"`
	OldState *OldState `json:"prev_content"`
}

func (e *OldState) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}
	return json.Marshal(e.Content)
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	ms := ts.UnixNano() / int64(time.Millisecond)
	return []byte(strconv.FormatInt(ms, 10)), nil
}

type GenericContent struct {
	Content   map[string]interface{}
	eventType EventType
}

func (c *GenericContent) EventType() EventType {
	return c.eventType
}

func (c *GenericContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Content)
}

type TestContent struct {
	Name string `json:"name"`
}

type MembershipEventContent struct {
	Membership  string `json:"membership"`
	DisplayName string `json:"displayname"`
	AvatarUrl   string `json:"avatar_url"`
}

func (c *MembershipEventContent) EventType() EventType {
	return EventTypeMembership
}

type CreateEventContent struct {
	Creator UserId `json:"creator"`
}

func (c *CreateEventContent) EventType() EventType {
	return EventTypeCreate
}

type NameEventContent struct {
	Name string `json:"name"`
}

func (c *NameEventContent) EventType() EventType {
	return EventTypeName
}

type TopicEventContent struct {
	Topic string `json:"topic"`
}

func (c *TopicEventContent) EventType() EventType {
	return EventTypeTopic
}

type AliasesEventContent struct {
	Aliases string `json:"aliases"`
}

func (c *AliasesEventContent) EventType() EventType {
	return EventTypeAliases
}

type PowerLevelsEventContent struct {
	Ban          int             `json:"ban"`
	Kick         int             `json:"kick"`
	Redact       int             `json:"redact"`
	UserDefault  int             `json:"users_default"`
	CreateState  int             `json:"state_default"`
	EventDefault int             `json:"events_default"`
	Users        map[UserId]int  `json:"users"`
	Events       map[EventId]int `json:"events"`
}

func (c *PowerLevelsEventContent) EventType() EventType {
	return EventTypePowerLevels
}

type JoinRulesEventContent struct {
	JoinRule JoinRule `json:"join_rule"`
}

func (c *JoinRulesEventContent) EventType() EventType {
	return EventTypeJoinRules
}