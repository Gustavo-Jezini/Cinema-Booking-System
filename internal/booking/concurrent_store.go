package booking

import "sync"

type ConcurrenceStore struct {
	bookings map[string]Booking
	sync.RWMutex
}

func NewConcurrenceStore() *ConcurrenceStore {
	return &ConcurrenceStore{
		bookings: map[string]Booking{},
	}
}

func (s *ConcurrenceStore) Book(b Booking) error {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.bookings[b.SeatID]; exists {
		return ErrSeatAlreadyBooked
	}

	s.bookings[b.SeatID] = b
	return nil
}

func (s *ConcurrenceStore) ListBookings(movieID string) []Booking {
	s.RLock()
	defer s.RUnlock()

	var result []Booking
	for _, b := range s.bookings {
		if b.MovieID == movieID {
			result = append(result, b)
		}
	}
	return result
}
