package controllers

import (
	"context"
	"golang-restaurant-management/database"
	"golang-restaurant-management/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var tableCollection = database.OpenCollection(database.Client, "table")

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := tableCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while finding all the table"})
			return
		}
		var allTables []bson.M
		if err = result.All(ctx, &allTables); err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, allTables)
	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var table models.Table
		var tableId = c.Param("table_id")
		err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching data from table Collection"})
			return
		}
		c.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var table models.Table
		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provide correct data for table booking"})
			return
		}

		validationErr := validate.Struct(table)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed for table data"})
			return
		}

		table.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.ID = primitive.NewObjectID()
		table.Table_id = table.ID.Hex()

		result, err := tableCollection.InsertOne(ctx, table)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed in inserting the table data"})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var table models.Table
		var updatedObj primitive.D
		var tableId = c.Param("table_id")
		filter := bson.M{"table_id": tableId}

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provide correct data for table booking"})
			return
		}

		if table.Table_number != nil {
			updatedObj = append(updatedObj, bson.E{Key: "table_number", Value: table.Table_number})
		}

		if table.Number_of_guests != nil {
			updatedObj = append(updatedObj, bson.E{Key: "number_of_guests", Value: table.Number_of_guests})
		}

		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		res, err := tableCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updatedObj},
			},
			&opt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update table request failed"})
			return
		}

		c.JSON(http.StatusOK, res)
	}
}
