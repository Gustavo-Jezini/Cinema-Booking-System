package booking

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	goredis "github.com/redis/go-redis/v9"

	"github.com/google/uuid"
)

func TestConcurrentBooking_ExactlyOneWins(t *testing.T) {
	rdb := goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}

	const seatKey = "seat:screen-1:A1"
	rdb.Del(context.Background(), seatKey)
	t.Cleanup(func() { rdb.Del(context.Background(), seatKey) })

	store := NewRedisStore(rdb)
	svc := NewService(store)

	const numGoroutines = 100_000 // 100k users trying to book a seat at the same time

	var (
		successes atomic.Int64
		failures  atomic.Int64
		wg        sync.WaitGroup
	)

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(userNum int) {
			defer wg.Done()
			_, err := svc.Book(Booking{
				MovieID: "screen-1",
				SeatID:  "A1",
				UserID:  uuid.New().String(),
			})
			if err == nil {
				successes.Add(1)
			} else {
				failures.Add(1)
			}
		}(i)
	}
	wg.Wait()

	if got := successes.Load(); got != 1 {
		t.Errorf("expected exactly 1 success, got %d", got)
	}
	if got := failures.Load(); got != int64(numGoroutines-1) {
		t.Errorf("expected %d failures, got %d", numGoroutines-1, got)
	}
}
