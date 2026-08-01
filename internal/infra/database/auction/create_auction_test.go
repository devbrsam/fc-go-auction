package auction

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestGetAuctionDuration(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected time.Duration
	}{
		{name: "configured duration", value: "250ms", expected: 250 * time.Millisecond},
		{name: "missing duration", value: "", expected: defaultAuctionDuration},
		{name: "invalid duration", value: "invalid", expected: defaultAuctionDuration},
		{name: "negative duration", value: "-1s", expected: defaultAuctionDuration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AUCTION_DURATION", tt.value)
			if got := GetAuctionDuration(); got != tt.expected {
				t.Errorf("GetAuctionDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAuctionAutomaticallyCloses(t *testing.T) {
	mongoURL := os.Getenv("MONGODB_TEST_URL")
	if mongoURL == "" {
		t.Skip("MONGODB_TEST_URL is not configured")
	}

	t.Setenv("AUCTION_DURATION", "500ms")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Fatalf("connect to MongoDB: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping MongoDB: %v", err)
	}

	databaseName := "auction_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	database := client.Database(databaseName)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_ = database.Drop(cleanupCtx)
		_ = client.Disconnect(cleanupCtx)
	})

	repository := NewAuctionRepository(database)
	auctionEntity, internalErr := auction_entity.CreateAuction(
		"Mechanical Keyboard",
		"Electronics",
		"A mechanical keyboard in excellent condition",
		auction_entity.New,
	)
	if internalErr != nil {
		t.Fatalf("create auction entity: %v", internalErr)
	}

	if internalErr := repository.CreateAuction(ctx, auctionEntity); internalErr != nil {
		t.Fatalf("persist auction: %v", internalErr)
	}

	storedAuction, internalErr := repository.FindAuctionById(ctx, auctionEntity.Id)
	if internalErr != nil {
		t.Fatalf("find open auction: %v", internalErr)
	}
	if storedAuction.Status != auction_entity.Active {
		t.Fatalf("initial status = %d, want %d", storedAuction.Status, auction_entity.Active)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		storedAuction, internalErr = repository.FindAuctionById(ctx, auctionEntity.Id)
		if internalErr != nil {
			t.Fatalf("find auction after expiration: %v", internalErr)
		}
		if storedAuction.Status == auction_entity.Completed {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("final status = %d, want %d", storedAuction.Status, auction_entity.Completed)
}
