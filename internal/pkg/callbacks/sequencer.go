package callbacks

import (
	"sync"

	"github.com/disgoorg/snowflake/v2"
)

// guildQueueBuffer bounds how many pending guild-scoped jobs may queue up
// behind a slow one. Sized well above realistic per-guild event bursts
// (channel switches are human-paced).
const guildQueueBuffer = 64

// guildSequencer serializes Discord role-mutating work per guild, so that
// VoiceStateUpdate and ChannelDelete events for the same guild are applied in
// the order Discord delivered them, while different guilds are still
// processed concurrently.
//
// Submit is called synchronously from disgo's single gateway read loop per
// shard, which is what makes the enqueue order match Discord's delivery
// order; the actual (potentially slow, REST-bound) work runs later on a
// per-guild worker goroutine, off that read loop, so it can't stall heartbeat
// processing. The zero value is ready to use.
type guildSequencer struct {
	mu     sync.Mutex
	queues map[snowflake.ID]chan func()
}

// Submit queues fn to run on guildID's dedicated worker, creating the worker
// on first use. fn runs after any previously submitted work for the same
// guild has completed.
//
// Submit never blocks: if the guild's queue is full, fn is dropped and Submit
// reports false. The caller holds disgo's event-manager mutex, so blocking
// here would wedge all event dispatch for the shard — including heartbeat
// ACK processing — until the queue drains (a production incident: a Discord
// role-create rate limit with a multi-hour retry_after backed up a guild
// queue and put the shard into a permanent zombie-reconnect loop). Dropping
// risks at most a stale ephemeral role, which the member's next voice event
// corrects.
func (s *guildSequencer) Submit(guildID snowflake.ID, fn func()) bool {
	select {
	case s.queue(guildID) <- fn:
		return true
	default:
		return false
	}
}

// SubmitWait queues fn like Submit but blocks until the guild's queue has
// capacity instead of dropping. It must not be called from the gateway read
// loop; it is for callers that can afford to wait (a goroutine falling back
// after a Submit drop, tests) while still needing fn to run serialized on
// the guild's worker.
func (s *guildSequencer) SubmitWait(guildID snowflake.ID, fn func()) {
	s.queue(guildID) <- fn
}

// Flush blocks until every job submitted for guildID before this call has
// completed.
func (s *guildSequencer) Flush(guildID snowflake.ID) {
	done := make(chan struct{})

	s.SubmitWait(guildID, func() { close(done) })

	<-done
}

// queue returns guildID's job channel, creating it and starting its drain
// worker on first use.
func (s *guildSequencer) queue(guildID snowflake.ID) chan func() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.queues == nil {
		s.queues = make(map[snowflake.ID]chan func())
	}

	queue, ok := s.queues[guildID]
	if !ok {
		queue = make(chan func(), guildQueueBuffer)
		s.queues[guildID] = queue

		go drainGuildQueue(queue)
	}

	return queue
}

func drainGuildQueue(queue chan func()) {
	for fn := range queue {
		fn()
	}
}
