package auction

import (
	"os"
	"time"
)

func GetAuctionDuration() time.Duration {
	auctionDuration := os.Getenv("AUCTION_DURATION")
	duration, err := time.ParseDuration(auctionDuration)
	if err != nil {
		return time.Minute * 5
	}

	return duration
}
