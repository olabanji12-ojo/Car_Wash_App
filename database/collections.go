package database


import (
	
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	UserCollection     *mongo.Collection
	CarCollection      *mongo.Collection
	CarwashCollection  *mongo.Collection
	BookingCollection  *mongo.Collection
	OrderCollection    *mongo.Collection
	ReviewCollection   *mongo.Collection
	PaymentCollection  *mongo.Collection
	ServiceCollection  *mongo.Collection
)

func InitCollections() {
	UserCollection = DB.Collection("users")
	CarCollection = DB.Collection("cars")
	CarwashCollection = DB.Collection("carwashes")
	BookingCollection = DB.Collection("bookings")
	OrderCollection = DB.Collection("orders")
	ReviewCollection = DB.Collection("reviews")
	PaymentCollection = DB.Collection("payments")
	ServiceCollection = DB.Collection("services")
}


