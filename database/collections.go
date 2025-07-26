package database


import (
	
	"go.mongodb.org/mongo-driver/mongo"

)

var (
	UserCollection         *mongo.Collection
	CarCollection          *mongo.Collection
	CarwashCollection      *mongo.Collection
	BookingCollection      *mongo.Collection
	OrderCollection        *mongo.Collection
	ReviewCollection       *mongo.Collection
	PaymentCollection      *mongo.Collection
	ServiceCollection      *mongo.Collection
	NotificationCollection *mongo.Collection
)

func InitCollections() { 
	UserCollection = DB.Collection("users") // touched
	CarCollection = DB.Collection("cars") // touched
	CarwashCollection = DB.Collection("carwashes") // touched
	BookingCollection = DB.Collection("bookings") // touched
	OrderCollection = DB.Collection("orders")
	ReviewCollection = DB.Collection("reviews")
	PaymentCollection = DB.Collection("payments")
	ServiceCollection = DB.Collection("services") // touched
	NotificationCollection = DB.Collection("notifications") // notifications
}


