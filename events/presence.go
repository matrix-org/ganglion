package events

import (
	"sync"
	"sync/atomic"

	"github.com/Rugvip/bullettime/interfaces"

	"github.com/Rugvip/bullettime/types"
)

type presenceStream struct {
	lock           sync.RWMutex
	events         map[types.UserId]indexedPresenceEvent
	max            uint64
	members        interfaces.MembershipStore
	asyncEventSink interfaces.AsyncEventSink
}

type indexedPresenceEvent struct {
	event types.PresenceEvent
	index uint64
}

func (m *indexedPresenceEvent) Event() types.Event {
	return &m.event
}

func (s *indexedPresenceEvent) Index() uint64 {
	return s.index
}

type updateFunc func(*types.User)

func NewPresenceStream(
	members interfaces.MembershipStore,
	asyncEventSink interfaces.AsyncEventSink,
) (interfaces.PresenceStream, error) {
	return &presenceStream{
		events:         map[types.UserId]indexedPresenceEvent{},
		members:        members,
		asyncEventSink: asyncEventSink,
	}, nil
}

func (s *presenceStream) SetUserProfile(userId types.UserId, profile types.UserProfile) (types.IndexedEvent, types.Error) {
	return s.update(userId, func(user *types.User) {
		user.UserProfile = profile
	})
}

func (s *presenceStream) SetUserStatus(userId types.UserId, status types.UserStatus) (types.IndexedEvent, types.Error) {
	return s.update(userId, func(user *types.User) {
		user.UserStatus = status
	})
}

func (s *presenceStream) update(userId types.UserId, updateFunc updateFunc) (types.IndexedEvent, types.Error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	indexed, existed := s.events[userId]
	if !existed {
		indexed.event.Content.UserId = userId
		indexed.event.EventType = types.EventTypePresence
	}
	updateFunc(&indexed.event.Content)
	index := atomic.AddUint64(&s.max, 1) - 1
	indexed.index = index
	s.events[userId] = indexed
	peerSet, err := s.members.Peers(userId)
	if err != nil {
		return nil, err
	}
	peers := make([]types.UserId, len(peerSet))
	for peer := range peerSet {
		peers = append(peers, peer)
	}
	s.asyncEventSink.Send(peers, &indexed)
	return &indexed, nil
}

func (s *presenceStream) Profile(user types.UserId) (types.UserProfile, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if indexed, ok := s.events[user]; ok {
		return indexed.event.Content.UserProfile, nil
	}
	return types.UserProfile{}, nil
}

func (s *presenceStream) Status(user types.UserId) (types.UserStatus, types.Error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if indexed, ok := s.events[user]; ok {
		return indexed.event.Content.UserStatus, nil
	}
	return types.UserStatus{}, nil
}

func (s *presenceStream) Max() uint64 {
	return atomic.LoadUint64(&s.max)
}

// Ignores user, roomSet, and limit
func (s *presenceStream) Range(
	user types.UserId,
	userSet map[types.UserId]struct{},
	roomSet map[types.RoomId]struct{},
	from, to uint64,
	limit uint,
) ([]types.IndexedEvent, types.Error) {
	var result []types.IndexedEvent
	if len(userSet) == 0 || from >= to {
		return result, nil
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	result = make([]types.IndexedEvent, 0, len(userSet))
	for user := range userSet {
		event := s.events[user]
		if event.index >= from && event.index < to {
			result = append(result, &event)
		}
	}
	return result, nil
}
