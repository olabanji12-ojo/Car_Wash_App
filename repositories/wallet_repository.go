package repositories

import (
	"context"
	"time"

	"github.com/olabanji12-ojo/CarWashApp/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WalletRepository struct {
	db *mongo.Database
}

func NewWalletRepository(db *mongo.Database) *WalletRepository {
	return &WalletRepository{db: db}
}

func (wr *WalletRepository) GetWalletByUserID(ctx context.Context, userID string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := wr.db.Collection("wallets").FindOne(ctx, bson.M{"user_id": userID}).Decode(&wallet)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (wr *WalletRepository) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	wallet.ID = primitive.NewObjectID()
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()
	_, err := wr.db.Collection("wallets").InsertOne(ctx, wallet)
	return err
}

func (wr *WalletRepository) UpdateBalance(ctx context.Context, userID string, amount float64) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$inc": bson.M{"balance": amount},
		"$set": bson.M{"updated_at": time.Now()},
	}
	
	// We use Upsert logic if wallet doesnt exist (though in escrow release we expect it to exist or be merchant)
	opts := options.Update().SetUpsert(true)
	_, err := wr.db.Collection("wallets").UpdateOne(ctx, filter, update, opts)
	return err
}

func (wr *WalletRepository) CreateTransaction(ctx context.Context, tx *models.WalletTransaction) error {
	tx.ID = primitive.NewObjectID()
	tx.CreatedAt = time.Now()
	_, err := wr.db.Collection("wallet_transactions").InsertOne(ctx, tx)
	return err
}

func (wr *WalletRepository) FindEscrowTransaction(ctx context.Context, bookingID string) (*models.WalletTransaction, error) {
	var tx models.WalletTransaction
	filter := bson.M{
		"metadata.booking_id": bookingID,
		"status":              "escrow",
	}
	err := wr.db.Collection("wallet_transactions").FindOne(ctx, filter).Decode(&tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (wr *WalletRepository) UpdateTransactionStatus(ctx context.Context, txID primitive.ObjectID, status string) error {
	_, err := wr.db.Collection("wallet_transactions").UpdateOne(
		ctx,
		bson.M{"_id": txID},
		bson.M{"$set": bson.M{"status": status}},
	)
	return err
}
