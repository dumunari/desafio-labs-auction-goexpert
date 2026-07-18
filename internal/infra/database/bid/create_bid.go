package bid

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/entity/bid_entity"
	"fullcycle-auction_go/internal/infra/database/auction"
	"fullcycle-auction_go/internal/internal_error"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type BidEntityMongo struct {
	Id        string  `bson:"_id"`
	UserId    string  `bson:"user_id"`
	AuctionId string  `bson:"auction_id"`
	Amount    float64 `bson:"amount"`
	Timestamp int64   `bson:"timestamp"`
}

type BidRepository struct {
	Collection            *mongo.Collection
	AuctionRepository     *auction.AuctionRepository
	auctionInterval       time.Duration
	auctionStatusMap      map[string]auction_entity.AuctionStatus
	auctionEndTimeMap     map[string]time.Time
	auctionStatusMapMutex *sync.Mutex
	auctionEndTimeMutex   *sync.Mutex
}

func NewBidRepository(database *mongo.Database, auctionRepository *auction.AuctionRepository) *BidRepository {
	return &BidRepository{
		auctionInterval:       auction.GetAuctionDuration(),
		auctionStatusMap:      make(map[string]auction_entity.AuctionStatus),
		auctionEndTimeMap:     make(map[string]time.Time),
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
		Collection:            database.Collection("bids"),
		AuctionRepository:     auctionRepository,
	}
}

func (bd *BidRepository) CreateBid(
	ctx context.Context,
	bidEntities []bid_entity.Bid) *internal_error.InternalError {
	for _, bid := range bidEntities {
		bd.auctionStatusMapMutex.Lock()
		auctionStatus, okStatus := bd.auctionStatusMap[bid.AuctionId]
		bd.auctionStatusMapMutex.Unlock()

		bd.auctionEndTimeMutex.Lock()
		auctionEndTime, okEndTime := bd.auctionEndTimeMap[bid.AuctionId]
		bd.auctionEndTimeMutex.Unlock()

		if !okStatus || !okEndTime {
			auctionEntity, err := bd.AuctionRepository.FindAuctionById(ctx, bid.AuctionId)
			if err != nil {
				logger.Error("Error trying to find auction by id", err)
				return err
			}

			auctionStatus = auctionEntity.Status
			auctionEndTime = auctionEntity.Timestamp.Add(bd.auctionInterval)

			bd.auctionStatusMapMutex.Lock()
			bd.auctionStatusMap[bid.AuctionId] = auctionStatus
			bd.auctionStatusMapMutex.Unlock()

			bd.auctionEndTimeMutex.Lock()
			bd.auctionEndTimeMap[bid.AuctionId] = auctionEndTime
			bd.auctionEndTimeMutex.Unlock()
		}

		now := time.Now()
		if auctionStatus == auction_entity.Closed || now.After(auctionEndTime) {
			return internal_error.NewBadRequestError("auction is closed")
		}

		bidEntityMongo := &BidEntityMongo{
			Id:        bid.Id,
			UserId:    bid.UserId,
			AuctionId: bid.AuctionId,
			Amount:    bid.Amount,
			Timestamp: bid.Timestamp.Unix(),
		}

		if _, err := bd.Collection.InsertOne(ctx, bidEntityMongo); err != nil {
			logger.Error("Error trying to insert bid", err)
			return internal_error.NewInternalServerError("Error trying to insert bid")
		}
	}

	return nil
}
