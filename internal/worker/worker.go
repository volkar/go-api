package worker

import (
	"context"
	"log/slog"
	"time"
)

type TokenRepository interface {
	Cleanup(ctx context.Context) error
}

func StartCleanupWorker(ctx context.Context, tokensRepo TokenRepository, logger *slog.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("Cleanup worker started with interval", "interval", interval)

	for {
		select {
		case <-ticker.C:
			// Separate context
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := tokensRepo.Cleanup(cleanupCtx)
			cancel()

			if err != nil {
				logger.Warn("Error during tokens cleanup", "error", err)
			} else {
				logger.Info("Expired tokens cleaned up successfully")
			}

		case <-ctx.Done():
			logger.Info("Cleanup worker stopping...")
			return
		}
	}
}
