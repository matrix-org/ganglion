package events

import (
	"sync"
	"sync/atomic"

	"github.com/Rugvip/bullettime/interfaces"

	"github.com/Rugvip/bullettime/types"
)

type typingSource struct {
	sync.RWMutex
	states    map[types.RoomId]*indexedTypingState
	max       uint64
	members   interfaces.MembershipStore
	eventSink interfaces.UserEventSink
}

type indexedTypingState struct {
	index uint64
	event types.TypingEvent
}

func NewTypingSource(
	members interfaces.MembershipStore,
	eventSink interfaces.UserEventSink,
) (typingSource, error) {
	return typingSource{
		states:    map[types.RoomId]*indexedTypingState{},
		members:   members,
		eventSink: eventSink,
	}, nil
}

func (s *typingSource) SetTyping(room types.RoomId, user types.UserId, typing bool) types.Error {
	s.Lock()
	defer s.Unlock()
	state := s.states[room]
	index := atomic.AddUint64(&s.max, 1) - 1
	if state == nil {
		state = &indexedTypingState{index: index}
		state.event.RoomId = room
		s.states[room] = state
	} else {
		state.index = index
	}
	userIds := state.event.Content.UserIds
	if typing {
		for _, member := range userIds {
			if member == user {
				return nil
			}
		}
		state.event.Content.UserIds = append(userIds, user)
	} else {
		for i, member := range userIds {
			if member == user {
				userIds[i] = userIds[len(userIds)-1]
				state.event.Content.UserIds = userIds[:len(userIds)-1]
				break
			}
		}
	}
	roomMembers, err := s.members.Users(room)
	if err != nil {
		return err
	}
	s.eventSink.Send(roomMembers, &state.event, index)
	return nil
}

func (s *typingSource) Typing(room types.RoomId) ([]types.UserId, types.Error) {
	s.RLock()
	defer s.RUnlock()
	state := s.states[room]
	if state == nil {
		return []types.UserId{}, nil
	}
	return state.event.Content.UserIds, nil
}

func (s *typingSource) Max() (uint64, types.Error) {
	return atomic.LoadUint64(&s.max), nil
}

func (s *typingSource) EventRange(userId types.UserId, from, to uint64) ([]types.Event, types.Error) {
	var result []types.Event
	rooms, err := s.members.Rooms(userId)
	if err != nil {
		return result, err
	}
	if len(rooms) == 0 || from >= to {
		return result, nil
	}
	s.RLock()
	defer s.RUnlock()
	result = make([]types.Event, len(rooms))
	for _, room := range rooms {
		state := s.states[room]
		if state.index >= from && state.index < to {
			result = append(result, &state.event)
		}
	}
	return result, nil
}