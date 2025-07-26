// MongoDB script to create geospatial index for location-based queries
// Run this in MongoDB Compass or mongo shell

// Create 2dsphere index on location field for carwashes collection
db.carwashes.createIndex({ "location": "2dsphere" });

// Verify the index was created
db.carwashes.getIndexes();

// Optional: Create compound index for better performance on filtered location queries
db.carwashes.createIndex({ 
    "location": "2dsphere", 
    "is_active": 1, 
    "has_location": 1 
});

console.log("Geospatial indexes created successfully!");
console.log("You can now use $near, $geoWithin, and other geospatial queries on the location field.");
