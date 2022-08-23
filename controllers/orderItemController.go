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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderItemPack struct {
	Table_id    *string
	Order_items []models.OrderItem
}

var orderItemCollection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		result, err := orderItemCollection.Find(context.TODO(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while finding all the orderitems"})
			return
		}
		var allOrderItems []bson.M
		if err = result.All(ctx, &allOrderItems); err != nil {
			log.Fatal(err)
			return
		}
		c.JSON(http.StatusOK, allOrderItems)
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var orderItem models.OrderItem
		var orderItemId = c.Param("orderItem_id")
		err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching data from orderItemCollection"})
			return
		}
		c.JSON(http.StatusOK, orderItem)
	}
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var orderItemPack OrderItemPack
		var order models.Order
		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		order.Order_Date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		orderItemsToBeInserted := []interface{}{}
		order.Table_id = orderItemPack.Table_id
		order_id := OrderItemOrderCreator(order)

		for _, orderItem := range orderItemPack.Order_items {
			orderItem.Order_id = order_id

			validationErr := validate.Struct(orderItem)
			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}
			orderItem.ID = primitive.NewObjectID()
			orderItem.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Order_item_id = orderItem.ID.Hex()
			var num = toFixed(*orderItem.Unit_price, 2)
			orderItem.Unit_price = &num
			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}
		result, err := orderItemCollection.InsertMany(ctx, orderItemsToBeInserted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "insertmany operation failed in orderItems"})
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, result)
	}

}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var orderItem models.OrderItem
		var updatedObj primitive.D
		var orderItemId = c.Param("orderItem_id")
		// binding the given update
		filter := bson.M{"order_item_id": orderItemId}

		if err := c.BindJSON(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": " binding orderItem faield"})
			return
		}

		if orderItem.Food_id != nil {
			updatedObj = append(updatedObj, bson.E{Key: "food_id", Value: orderItem.Food_id})
		}

		if orderItem.Quantity != nil {
			updatedObj = append(updatedObj, bson.E{Key: "quantity", Value: orderItem.Quantity})
		}

		if orderItem.Unit_price != nil {
			updatedObj = append(updatedObj, bson.E{Key: "unit_price", Value: orderItem.Unit_price})
		}

		orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updatedObj = append(updatedObj, bson.E{Key: "updated_at", Value: orderItem.Updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updatedObj},
			},
			&opt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed for order Item request"})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderId = c.Param("order_id")
		allOrdersItems, err := ItemsByOrder(orderId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while lising items by the order id"})
			return
		}
		c.JSON(http.StatusOK, allOrdersItems)
	}
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "order_id", Value: id}}}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "food"}, {Key: "localField", Value: "food_id"}, {Key: "foreignField", Value: "food_id"}, {Key: "as", Value: "food"}}}}
	upwindStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$food"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	// lookupfor order
	lookupOrderStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "order"}, {Key: "localField", Value: "order_id"}, {Key: "foreignField", Value: "order_id"}, {Key: "as", Value: "order"}}}}
	upwindOrderStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$order"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	// lookupfor table
	lookupTableStage := bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "table"}, {Key: "localField", Value: "order.table_id"}, {Key: "foreignField", Value: "table_id"}, {Key: "as", Value: "table"}}}}
	unwindTableStage := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$table"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}}

	projectState := bson.D{
		{
			Key: "$project", Value: bson.D{
				{Key: "id", Value: 0},
				{Key: "amount", Value: "$food.price"},
				{Key: "total_count", Value: 1},
				{Key: "food_name", Value: "$food.name"},
				{Key: "food_image", Value: "$food.food_image"},
				{Key: "table_number", Value: "$table.table_number"},
				{Key: "table_id", Value: "$table.table_id"},
				{Key: "order_id", Value: "$order.order_id"},
				{Key: "price", Value: "$food.price"},
				{Key: "quantity", Value: 1},
			},
		},
	}

	groupStage := bson.D{
		{
			Key: "$group", Value: bson.D{
				{Key: "id", Value: bson.D{
					{
						Key: "order_id", Value: "$order_id",
					},
					{
						Key: "table_id", Value: "$table_id",
					},
					{
						Key: "table_number", Value: "$table_number",
					}},
				},
				{
					Key: "payment_due", Value: bson.D{{Key: "$sum", Value: "$amount"}},
				},
				{
					Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}},
				},
				{
					Key: "order_items", Value: bson.D{{Key: "$push", Value: "$$ROOT"}},
				},
			},
		},
	}

	projectState2 := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "id", Value: 0},
			{Key: "payment_due", Value: 1},
			{Key: "total_count", Value: 1},
			{Key: "table_number", Value: "$_id.table_number"},
			{Key: "order_items", Value: 1},
		}},
	}

	result, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
		upwindStage,
		lookupOrderStage,
		upwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		projectState,
		groupStage,
		projectState2,
	})
	if err != nil {
		panic(err)
	}
	if err := result.All(ctx, &OrderItems); err != nil {
		panic(err)
	}

	return OrderItems, err
}
