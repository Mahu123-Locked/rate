package rate

import (
	"fmt"
	"time"

	"github.com/webriots/rate/time56"
)

// TokenBucketLimiter implements the token bucket algorithm for rate
// limiting. It maintains multiple buckets to distribute load and
// reduce contention. Each bucket has a fixed capacity and refills at
// a specified rate.
type TokenBucketLimiter struct {
	buckets             atomicSliceUint64 // Array of token buckets
	bucketMask          uint              // Bit mask for IDs to buckets
	burstCapacity       uint8             // Maximum tokens per bucket
	refillIntervalNanos int64             // Nanoseconds per token refill
	numBuckets          uint              // Number of buckets (pow^2)
	refillRate          float64           // Original refill rate value
	refillRateUnit      time.Duration     // Time unit for refill rate
}

// NewTokenBucketLimiter creates a new token bucket rate limiter with
// the specified parameters:
//
//   - numBuckets: number of token buckets (must be a power of two)
//   - burstCapacity: maximum number of tokens that can be consumed at
//     once
//   - refillRate: rate at which tokens are refilled
//   - refillRateUnit: time unit for refill rate calculations (e.g.,
//     time.Second)
//
// Returns a new TokenBucketLimiter instance and any error that
// occurred during creation. The numBuckets parameter must be a power
// of two for efficient hashing.
func NewTokenBucketLimiter(
	numBuckets uint,
	burstCapacity uint8,
	refillRate float64,
	refillRateUnit time.Duration,
) (*TokenBucketLimiter, error) {
	if powerOfTwo := (numBuckets != 0) && ((numBuckets & (numBuckets - 1)) == 0); !powerOfTwo {
		return nil, fmt.Errorf("numBuckets must be a power of two")
	}

	now := nowfn().UnixNano()
	stamp := time56.Unix(now)
	bucket := newTokenBucket(burstCapacity, stamp)
	packed := bucket.packed()
	buckets := newAtomicSliceUint64(int(numBuckets))
	for i := range numBuckets {
		buckets.Set(int(i), packed)
	}

	return &TokenBucketLimiter{
		burstCapacity:       burstCapacity,
		numBuckets:          numBuckets,
		refillRate:          refillRate,
		refillRateUnit:      refillRateUnit,
		refillIntervalNanos: nanoRate(refillRateUnit, refillRate),
		bucketMask:          numBuckets - 1,
		buckets:             buckets,
	}, nil
}

// Check returns whether a token would be available for the given ID
// without actually taking it. This is useful for preemptively
// checking if an operation would be rate limited before attempting
// it. Returns true if a token would be available, false otherwise.
func (t *TokenBucketLimiter) Check(id []byte) bool {
	index := t.index(id)
	return t.checkInner(index, t.refillIntervalNanos)
}

// TakeToken attempts to take a token for the given ID. It returns
// true if a token was successfully taken, false if the operation
// should be rate limited. This method is thread-safe and can be
// called concurrently from multiple goroutines.
func (t *TokenBucketLimiter) TakeToken(id []byte) bool {
	index := t.index(id)
	return t.takeTokenInner(index, t.refillIntervalNanos)
}

// checkInner is an internal method that checks if a token is
// available in the bucket at the specified index using the given
// refill rate. This is used by Check and is also used by other
// limiters that wrap this one.
func (t *TokenBucketLimiter) checkInner(index int, rate int64) bool {
	now := nowfn().UnixNano()
	existing := t.buckets.Get(index)
	bucket := unpack(existing)
	refilled := bucket.refill(now, rate, t.burstCapacity)
	return refilled.level > 0
}

// takeTokenInner is an internal method that attempts to take a token
// from the bucket at the specified index using the given refill rate.
// This is used by TakeToken and is also used by other limiters that
// wrap this one. It uses atomic operations to ensure thread safety.
func (t *TokenBucketLimiter) takeTokenInner(index int, rate int64) bool {
	now := nowfn().UnixNano()
	for {
		existing := t.buckets.Get(index)
		unpacked := unpack(existing)
		refilled := unpacked.refill(now, rate, t.burstCapacity)
		updated, taken := refilled.take()

		if updated != unpacked && !t.buckets.CompareAndSwap(
			index,
			existing,
			updated.packed(),
		) {
			continue
		}

		return taken
	}
}

// index calculates the bucket index for the given ID using the FNV-1a
// hash. The result is masked to ensure it falls within the range of
// valid buckets.
func (t *TokenBucketLimiter) index(id []byte) int {
	h := uint(14695981039346656037)
	for _, b := range id {
		h ^= uint(b)
		h *= 1099511628211
	}
	return int(h & t.bucketMask)
}

// tokenBucket represents a single token bucket with a certain level
// (number of tokens) and a timestamp of when it was last refilled.
type tokenBucket struct {
	level uint8       // Current number of tokens in the bucket
	stamp time56.Time // Last time the bucket was refilled
}

// newTokenBucket creates a new token bucket with the specified level
// and timestamp.
func newTokenBucket(level uint8, stamp time56.Time) tokenBucket {
	return tokenBucket{level: level, stamp: stamp}
}

// refill updates the token bucket based on elapsed time since the
// last refill. It calculates how many tokens should be added based on
// the elapsed time and refill rate, and updates the bucket's level
// and timestamp accordingly. The bucket level will not exceed
// maxLevel.
func (b tokenBucket) refill(nowNS, rate int64, maxLevel uint8) tokenBucket {
	now := time56.Unix(nowNS)

	elapsed := now.Since(b.stamp)
	if elapsed <= 0 {
		return b
	}

	tokens := elapsed / rate
	if tokens <= 0 {
		return b
	}

	level := maxLevel
	if avail := maxLevel - b.level; tokens < int64(avail) {
		level = b.level + uint8(tokens)
	}

	if b.level != level {
		remainder := elapsed % rate
		b.stamp = now.Sub(remainder)
		b.level = level
	}

	return b
}

// take attempts to take a token from the bucket. Returns the updated
// bucket and a boolean indicating whether a token was taken. If no
// tokens are available, the bucket remains unchanged and false is
// returned.
func (b tokenBucket) take() (tokenBucket, bool) {
	if b.level > 0 {
		b.level--
		return b, true
	} else {
		return b, false
	}
}

// packed converts the token bucket to a packed uint64 representation
// where the level is stored in the high 8 bits and the timestamp in
// the low 56 bits.
func (b tokenBucket) packed() uint64 {
	return b.stamp.Pack(b.level)
}

// unpack extracts a token bucket from its packed uint64
// representation. This is the inverse operation of packed().
func unpack(packed uint64) tokenBucket {
	return newTokenBucket(time56.Unpack(packed))
}

// nanoRate converts a refill rate from tokens per unit to nanoseconds
// per token. This is used to calculate how frequently tokens should
// be added to buckets.
func nanoRate(refillRateUnit time.Duration, refillRate float64) int64 {
	return int64(float64(refillRateUnit.Nanoseconds()) * refillRate)
}
