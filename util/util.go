package util

import (
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"golang.org/x/time/rate"
)

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

// func GetLimiter(ip string) *rate.Limiter {
// 	mu.Lock()
// 	defer mu.Unlock()

//		limiter, exists := limiters[ip]
//		if !exists {
//			// 1 request per 30 seconds, with burst of 1
//			limiter = rate.NewLimiter(rate.Every(30*time.Second), 1)
//			limiters[ip] = limiter
//		}
//		return limiter
//	}
func GetLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := limiters[ip]
	if !exists {
		// 5 events per minute (i.e., 1 event every 12 seconds), burst of 5
		limiter = rate.NewLimiter(rate.Every(12*time.Second), 5)
		limiters[ip] = limiter
	}
	return limiter
}

func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code.Name() == "unique_violation" {
			return true
		}
	}
	if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return true
	}
	return false
}

