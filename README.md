# Cinema Booking System

> A distributed seat reservation system for cinema theaters that prevents double-booking through pessimistic locking with Redis.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 📋 Overview

The Cinema Booking System is a backend service that handles concurrent seat reservations for movie theaters. It solves the critical race condition problem where multiple users attempt to purchase the same seat simultaneously.

**The Core Problem:**
- User A clicks seat A1 at 10:00:00.000
- User B clicks the same seat A1 at 10:00:00.001
- Without proper synchronization, both could receive confirmation of purchase

**The Solution:**
This project implements a **pessimistic locking strategy** using Redis to guarantee atomicity. Only one user can successfully reserve any given seat, even under high concurrent load.

## 🎯 Key Features

- **Atomic Seat Reservation** - Uses Redis `SetNX` for guaranteed single ownership
- **Session Management** - Temporary "held" reservations with 2-minute auto-expiry
- **Distributed Architecture** - Works across multiple server instances
- **RESTful API** - Clean HTTP endpoints for all operations
- **Multiple Storage Backends** - Redis (production), Memory (testing)
- **User Authorization** - Ensures only session owners can confirm/cancel
- **Thread-Safe** - Handles goroutine concurrency safely

## 🏗️ Architecture

### Layered Design Pattern

```
HTTP Client
    ↓
┌─────────────────────────────────┐
│  Handler Layer                  │  ← HTTP routing & validation
├─────────────────────────────────┤
│  Service Layer                  │  ← Business logic orchestration
├─────────────────────────────────┤
│  BookingStore Interface         │  ← Data persistence contract
├─────────────────────────────────┤
│  Redis/Memory Implementation    │  ← Actual data storage
└─────────────────────────────────┘
    ↓
  Database
```

### Data Flow

```
Request → Handler validates → Service processes → Store persists
   ↓         (user_id, params)   (delegates)    (Redis atomic ops)
Response ← Handler formats ← Service returns ← Store confirms
```

### Redis Data Structure

```
Seat Reservations:
  seat:{movieID}:{seatID} = Booking JSON (with TTL when "held")

Session Lookups (reverse index):
  session:{sessionID} = seat:{movieID}:{seatID}

Example:
  seat:blood:A1 = {"id":"uuid-123","user_id":"user123","status":"held"}
  [TTL: 2 minutes]
  
  session:uuid-123 = "seat:blood:A1"
  [TTL: 2 minutes]
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+**
- **Redis 6.0+**
- **Docker & Docker Compose** (optional, for Redis)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Gustavo-Jezini/cinema-booking-system.git
   cd cinema-booking-system
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Start Redis** (using Docker Compose)
   ```bash
   docker-compose up -d
   ```
   Redis will be available at `localhost:6379`

4. **Run the application**
   ```bash
   go run ./cmd/main.go
   ```
   The server will start at `http://localhost:8080`

5. **Access the UI**
   Open your browser to `http://localhost:8080`

## 📡 API Endpoints

### Get Available Movies

```http
GET /movies
```

**Response:**
```json
[
  {
    "id": "blood",
    "title": "There Will be Blood",
    "rows": 5,
    "seats_per_row": 8
  },
  {
    "id": "budapest",
    "title": "The Great Hotel: Budapest",
    "rows": 4,
    "seats_per_row": 6
  }
]
```

---

### List Seats for a Movie

```http
GET /movies/{movieID}/seats
```

**Example:**
```bash
curl http://localhost:8080/movies/blood/seats
```

**Response:**
```json
[
  {
    "seat_id": "A1",
    "user_id": "user123",
    "booked": true,
    "confirmed": true
  },
  {
    "seat_id": "B3",
    "user_id": "user456",
    "booked": true,
    "confirmed": false
  }
]
```

---

### Hold a Seat (Reserve Temporarily)

```http
POST /movies/{movieID}/seats/{seatID}/hold
Content-Type: application/json

{
  "user_id": "user123"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/movies/blood/seats/A1/hold \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user123"}'
```

**Success Response (201 Created):**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "movieID": "blood",
  "seat_id": "A1",
  "expires_at": "2026-04-27T16:30:00Z"
}
```

**Error Response (409 Conflict - Seat Taken):**
```json
{
  "error": "seat is already taken"
}
```

**⚠️ Important:** The reservation automatically expires after **2 minutes** if not confirmed.

---

### Confirm Seat Purchase

```http
PUT /sessions/{sessionID}/confirm
Content-Type: application/json

{
  "user_id": "user123"
}
```

**Example:**
```bash
curl -X PUT http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/confirm \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user123"}'
```

**Success Response (200 OK):**
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "movie_id": "blood",
  "seat_id": "A1",
  "user_id": "user123",
  "status": "confirmed"
}
```

**Error Response (422 Unprocessable Entity - Wrong User):**
```json
{
  "error": "user does not own this session"
}
```

---

### Cancel Reservation

```http
DELETE /sessions/{sessionID}
Content-Type: application/json

{
  "user_id": "user123"
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user123"}'
```

**Success Response (204 No Content)**
- No body returned
- Seat is now available for other users

**Error Response (422 Unprocessable Entity - Wrong User):**
```json
{
  "error": "user does not own this session"
}
```

---

## 📊 State Machine

```
[Available]
    ↓ POST /hold (SetNX succeeds)
[Held] (TTL: 2 minutes)
    ├─ PUT /confirm → [Confirmed] (permanent)
    └─ DELETE / (TTL expires) → [Available]

[Confirmed]
    └─ DELETE / (manual release) → [Available]
```

## 🛠️ Project Structure

```
cinema-booking-system/
├── cmd/
│   └── main.go                      # Application entry point
├── internal/
│   ├── booking/
│   │   ├── domain.go                # Core types (Booking, BookingStore interface)
│   │   ├── handler.go               # HTTP handlers (Controller layer)
│   │   ├── service.go               # Business logic (Service layer)
│   │   ├── redis_store.go           # Redis implementation (production)
│   │   ├── memory_store.go          # In-memory implementation (testing)
│   │   ├── concurrent_store.go      # Thread-safe memory implementation
│   │   └── service_test.go          # Unit tests
│   ├── adapters/
│   │   └── redis/
│   │       └── redis.go             # Redis client configuration
│   └── utils/
│       └── utils.go                 # Helper functions (JSON serialization)
├── static/                          # HTML/CSS/JS frontend
├── docker-compose.yml               # Redis container setup
├── go.mod / go.sum                  # Dependency management
└── README.md / explanation.md       # Documentation
```

## 🧪 Testing

### Run Tests

```bash
go test ./...
```

### Run Tests with Verbose Output

```bash
go test -v ./...
```

### Run Specific Test

```bash
go test -run TestBooking ./internal/booking
```

### Test Coverage

```bash
go test -cover ./...
```

## 🔄 How It Works: Detailed Example

### Scenario: Two users try to book seat A1 simultaneously

**Timeline:**
```
T=0:00
User Alice: POST /movies/blood/seats/A1/hold {"user_id":"alice"}
User Bob:   POST /movies/blood/seats/A1/hold {"user_id":"bob"}
            (arrive within milliseconds)

T=0:01 (Handler processing)
Both requests extracted: movieID="blood", seatID="A1"

T=0:02 (Redis execution - ATOMIC)
Alice: Redis SET "seat:blood:A1" {...} NX EX 120  → "OK" ✓
Bob:   Redis SET "seat:blood:A1" {...} NX EX 120  → NULL ✗

T=0:03 (Response)
Alice: HTTP 201 Created + session_id
Bob:   HTTP 409 Conflict + "seat is already taken"

T=2:01 (If Alice doesn't confirm)
Redis TTL expires on "seat:blood:A1"
Seat is automatically released

T=2:02
Bob can now successfully POST /hold/A1
```

## 🔐 Security Features

- **User Authorization** - Only the session owner can confirm/cancel their reservation
- **Atomic Operations** - Redis guarantees no data races
- **Automatic Cleanup** - Held seats expire automatically
- **No Overbooking** - SetNX prevents duplicate seat assignments

## 🚀 Performance Considerations

| Operation | Time Complexity | Notes |
|-----------|-----------------|-------|
| Hold Seat | O(1) | Single Redis SET operation |
| List Seats | O(n) | Scans all seats for a movie |
| Confirm | O(1) | Direct Redis operations |
| Cancel | O(1) | Direct Redis operations |

**Redis Performance:**
- ~1000s requests/second on commodity hardware
- Sub-millisecond latency for SET/GET
- Built-in persistence with RDB/AOF

## 📖 Additional Documentation

- **[Detailed Explanation](./explanation.md)** - In-depth walkthrough of the architecture, data structures, and request flow

## 🛠️ Technologies Used

| Technology | Purpose |
|------------|---------|
| **Go 1.21+** | Backend language |
| **Redis 6.0+** | Distributed locking & data store |
| **Docker** | Redis containerization |
| **Standard Library** | HTTP, JSON, UUID generation |

## 📝 Example Workflow: Complete User Journey

```bash
# 1. List available movies
curl http://localhost:8080/movies
# Response: [{"id":"blood","title":"There Will be Blood",...}]

# 2. Check available seats
curl http://localhost:8080/movies/blood/seats
# Response: [{"seat_id":"A1","booked":false,...}]

# 3. Hold a seat (temporary reservation)
curl -X POST http://localhost:8080/movies/blood/seats/A1/hold \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user123"}'
# Response: {"session_id":"uuid-123","expires_at":"2026-04-27T16:30:00Z"}

# 4. Confirm the purchase (within 2 minutes)
curl -X PUT http://localhost:8080/sessions/uuid-123/confirm \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user123"}'
# Response: {"session_id":"uuid-123","status":"confirmed"}

# 5. Check seats again - now A1 shows as confirmed
curl http://localhost:8080/movies/blood/seats
# Response: [{"seat_id":"A1","user_id":"user123","booked":true,"confirmed":true}]
```

## 🎓 Learning Outcomes

This project demonstrates:
- ✅ Distributed systems synchronization
- ✅ Race condition prevention
- ✅ Redis atomic operations
- ✅ Layered architecture design
- ✅ RESTful API design
- ✅ Dependency injection pattern
- ✅ Concurrent programming in Go
- ✅ Interface-driven development

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👤 Author

**Gustavo Jezini**

## 🤝 Contributing

This is an educational project. Feel free to fork and experiment with your own improvements!

## ❓ FAQ

### Why Redis and not just a mutex?
A mutex only works on a single server. Redis works across multiple servers, which is necessary for production systems with load balancing.

### What happens after 2 minutes?
Held seats automatically expire in Redis. The user must confirm before the TTL expires or start over.

### Can a user hold multiple seats?
Yes, each seat has its own reservation. A user can hold seat A1 and B3 simultaneously with different session IDs.

### What about seat confirmation without holding first?
The API enforces the flow: you must hold first, then confirm. There's no way to skip the hold step.

### How does this scale?
Redis can handle thousands of operations per second. For massive scale, consider Redis cluster mode or sharding by movieID.

---

**For more detailed information about how the system works, see [explanation.md](./explanation.md)**
