package bucket

import (
	"loadbalancer/internal/config"
	"sync"
	"time"
)

type BucketManager struct {
	config        *config.Config
	mux           sync.Mutex
	buckets       map[string]*TokenBucket
	stopCleanup   chan struct{} // канал для остановки горутины отчистки
	ipToRateLimit map[string]struct {
		RequestsPerSec int
		Burst          int
	}
}

// NewBucketManager - конструктор BucketManager
func NewBucketManager(cfg *config.Config) *BucketManager {
	bm := &BucketManager{
		config:      cfg,
		buckets:     make(map[string]*TokenBucket),
		stopCleanup: make(chan struct{}),
		ipToRateLimit: make(map[string]struct {
			RequestsPerSec int
			Burst          int
		}),
	}

	// заполняем экземпляр бакет менеджера
	for _, specialLimit := range cfg.RateLimit.SpecialLimits {
		for _, ip := range specialLimit.IPs {
			bm.ipToRateLimit[ip] = struct {
				RequestsPerSec int
				Burst          int
			}{
				RequestsPerSec: specialLimit.Limit.RequestsPerSec,
				Burst:          specialLimit.Limit.Burst,
			}
		}
	}

	if cfg.RateLimit.Enabled {
		bm.startCleanupRoutine()
	}

	return bm
}

// startCleanupRoutine - горутина для запуска отчистки старых бакетов
func (bm *BucketManager) startCleanupRoutine() {
	ticker := time.NewTicker(bm.config.RateLimit.CleanupInterval * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				bm.cleanupOldBuckets()
			case <-bm.stopCleanup:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop - останавливает горутину с отчисткой бакетов
func (bm *BucketManager) Stop() {
	if bm.config.RateLimit.Enabled {
		close(bm.stopCleanup)
	}
}

// cleanupOldBuckets - удаляет бакеты с временем последнего обращения старше чем cleanupInterval
func (bm *BucketManager) cleanupOldBuckets() {
	bm.mux.Lock()
	defer bm.mux.Unlock()

	minusInterval := -bm.config.RateLimit.CleanupInterval * time.Second
	cutoff := time.Now().Add(minusInterval)
	for ip, bucket := range bm.buckets {
		bucket.mux.Lock()
		if bucket.lastCheckTime.Before(cutoff) {
			delete(bm.buckets, ip)
		}
		bucket.mux.Unlock()
	}
}

// Allow -
func (bm *BucketManager) Allow(ip string) bool {
	// если ограничитель выключен, то всегда даём добро на все запросы
	if !bm.config.RateLimit.Enabled {
		return true
	}

	// получаем дефолтные лимиты
	requestsPerSec := bm.config.RateLimit.Default.RequestsPerSec
	burst := bm.config.RateLimit.Default.Burst

	// получаем лимиты для IP, если есть
	if limit, ok := bm.ipToRateLimit[ip]; ok {
		requestsPerSec = limit.RequestsPerSec
		burst = limit.Burst
	}

	bm.mux.Lock()
	defer bm.mux.Unlock()

	bucket, exists := bm.buckets[ip]
	if !exists { // новый IP
		bucket = NewTokenBucket(requestsPerSec, burst)
		bm.buckets[ip] = bucket
	}

	return bucket.Allow()
}
