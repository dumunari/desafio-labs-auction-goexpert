package auction

import (
	"context"
	"os"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCreateAuction_ExpiresAutomatically(t *testing.T) {
	t.Setenv("AUCTION_DURATION", "1s")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	db, cleanup := newAuctionTestDatabase(t, ctx)
	defer cleanup()

	repo := NewAuctionRepository(db)

	auctionEntity, internalErr := auction_entity.CreateAuction(
		"iPhone 15",
		"electronics",
		"Aparelho em excelente estado de conservacao para teste de expiracao.",
		auction_entity.New,
	)
	if internalErr != nil {
		t.Fatalf("unexpected internal error creating auction entity: %v", internalErr)
	}

	if internalErr := repo.CreateAuction(context.Background(), auctionEntity); internalErr != nil {
		t.Fatalf("unexpected internal error creating auction in repository: %v", internalErr)
	}

	createdAuction, internalErr := repo.FindAuctionById(context.Background(), auctionEntity.Id)
	if internalErr != nil {
		t.Fatalf("unexpected internal error finding created auction: %v", internalErr)
	}
	if createdAuction.Status != auction_entity.Active {
		t.Fatalf("expected auction status to be Active right after creation, got %v", createdAuction.Status)
	}

	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		updatedAuction, findErr := repo.FindAuctionById(context.Background(), auctionEntity.Id)
		if findErr == nil && updatedAuction.Status == auction_entity.Closed {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	finalAuction, finalErr := repo.FindAuctionById(context.Background(), auctionEntity.Id)
	if finalErr != nil {
		t.Fatalf("auction did not expire and could not be loaded in final check: %v", finalErr)
	}

	t.Fatalf("expected auction status to become Closed after expiration, got %v", finalAuction.Status)
}

func newAuctionTestDatabase(t *testing.T, ctx context.Context) (*mongo.Database, func()) {
	t.Helper()

	mongoURL := os.Getenv("AUCTION_TEST_MONGODB_URL")
	if mongoURL == "" {
		mongoURL = os.Getenv("MONGODB_URL")
	}
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@localhost:27017/?authSource=admin"
	}

	databaseName := os.Getenv("AUCTION_TEST_MONGODB_DB")
	if databaseName == "" {
		databaseName = os.Getenv("MONGODB_DB")
	}
	if databaseName == "" {
		databaseName = "auctions"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Skipf("skipping test: unable to connect to MongoDB (%v)", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		t.Skipf("skipping test: unable to ping MongoDB (%v)", err)
	}

	database := client.Database(databaseName)
	collection := database.Collection("auctions")
	_, _ = collection.DeleteMany(ctx, map[string]any{})

	cleanup := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()

		_, _ = collection.DeleteMany(cleanupCtx, map[string]any{})
		_ = client.Disconnect(cleanupCtx)
	}

	return database, cleanup
}
