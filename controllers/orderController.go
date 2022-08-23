package controllers

import (
	"context"
	"golang-restaurant-management/database"
	"golang-restaurant-management/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var orderCollection = database.OpenCollection(database.Client,"order")

func GetOrders() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		orderCursor, err := orderCollection.Find(ctx,bson.M{})
		if(err != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":" error while querying for the orders"})
			return
		}
		var allOrders []primitive.M
		if err := orderCursor.All(ctx,&allOrders); err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		c.JSON(http.StatusOK,allOrders)
	}
}

func GetOrder() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		var order models.Order
		var orderId = c.Param("order_id")
		
		err := orderCollection.FindOne(ctx,bson.M{"order_id":orderId}).Decode(&order)
		if(err != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":"errot in fetching the order"})
			return
		}
		c.JSON(http.StatusOK,order)
	}
}

func CreateOrder() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		var table models.Table
		var order models.Order

		if err := c.BindJSON(&order); err!= nil {
			c.JSON(http.StatusBadRequest,gin.H{"error" : err.Error()})
			return
		}

		ValidateErr := validate.Struct(order)
		if ValidateErr != nil {
			c.JSON(http.StatusBadRequest,gin.H{"error":ValidateErr.Error()})
			return
		}
		if(order.Table_id != nil){
			err := tableCollection.FindOne(ctx,bson.M{"table_id":order.Table_id}).Decode(&table)
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"Menu does not exist"})
				return
			}
		}else{
			c.JSON(http.StatusBadRequest,gin.H{"error":"Table id is required"})
			return
		}

		order.Created_at, _ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()
	
		result, inserterr := orderCollection.InsertOne(ctx,order)
		if(inserterr != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":inserterr.Error()})
			return
		}

		c.JSON(http.StatusOK,result)
	}
}

func UpdateOrder() gin.HandlerFunc{
	return func(c *gin.Context) {
		var table models.Table
		var order models.Order

		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		var orderId = c.Param("order_id")
		var updateObj bson.D
		if err := c.BindJSON(&order); err != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":"failed during binding order"})
			return
		}
		
		if order.Table_id != nil{
			err := tableCollection.FindOne(ctx, bson.M{"table_id":order.Table_id}).Decode(&table)
			if(err != nil){
				c.JSON(http.StatusInternalServerError,gin.H{"error":"failed during fetching table id for order"})
				return
			}

			updateObj = append(updateObj, bson.E{Key: "table_id",Value: order.Table_id})
		}

		order.Updated_at, _ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at",Value: order.Updated_at})

		upsert := true
		filter := bson.M{"order_id":orderId}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set",Value: updateObj},
			},
			&opt,
		)
		if(err != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":"order update failed"})
			return
		}
		c.JSON(http.StatusOK,result)
	}
}

func OrderItemOrderCreator(order models.Order) string{
	var ctx,cancel = context.WithTimeout(context.Background(),10*time.Second)
	order.Created_at, _ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(ctx,order)
	defer cancel()
	return order.Order_id
}