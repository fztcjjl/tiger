package uuid

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"
)

type UUID struct {
	t time.Time
	r int64
}

func (u *UUID) String() string {
	v := fmt.Sprintf("%d-%d", u.t.UnixNano(), u.r)
	hash := fnv.New64a()
	hash.Write([]byte(v))
	return fmt.Sprintf("%d", hash.Sum64())
}

func New() *UUID {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &UUID{
		t: time.Now(),
		r: r.Int63(),
	}
}
