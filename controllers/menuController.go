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

var menuCollection = database.OpenCollection(database.Client,"menu")

func GetMenus() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		result, err := menuCollection.Find(context.TODO(),bson.M{})
		defer cancel()
		if(err != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}
		var allMenus []bson.M
		if err = result.All(ctx,&allMenus); err!=nil{
			log.Fatal(err)
		}
		c.JSON(http.StatusOK,allMenus)
	}
}

func GetMenu() gin.HandlerFunc{
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		var menu models.Menu;
		var menuId = c.Param("menu_id")
		err := menuCollection.FindOne(ctx,bson.M{"menu_id":menuId}).Decode(&menu)
		if( err != nil) {
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while fetching data from menuCollection"})
			return
		}
		c.JSON(http.StatusOK,menu)
	}
}

func CreateMenu() gin.HandlerFunc{
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		var menu models.Menu

		if err:=c.BindJSON(&menu); err != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return
		}

		validateMenu := validate.Struct(menu)
		if validateMenu != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":validateMenu.Error()})
			return
		}

		menu.Created_at, _ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		menu.Updated_at, _ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		menu.ID = primitive.NewObjectID()
		menu.Menu_id = menu.ID.Hex()
		// menu object is ready
		// now time to insert into the collection
		result,err := menuCollection.InsertOne(ctx,menu)
		if(err != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}
		c.JSON(http.StatusOK,result)
		defer cancel()
	}
}

func inTimeSpan(Start_Date time.Time, End_Date time.Time, Current_time time.Time) (bool){
	return Start_Date.After(time.Now()) && End_Date.After(Start_Date)
}

func UpdateMenu() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		// need to check if menu with this menu id exist or not
		var menu models.Menu

		if err := c.BindJSON(&menu);err != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return
		}
		menuId := c.Param("menu_id")
		filter := bson.M{"menu_id":menuId}
		var UpdateObj primitive.D

		if(menu.Start_Date != nil && menu.End_date != nil){

			if(!inTimeSpan(*menu.Start_Date,*menu.End_date,time.Now())){
				c.JSON(http.StatusInternalServerError,gin.H{"error":"kindly enter correct dates"})
				defer cancel()
				return
			}
			UpdateObj = append(UpdateObj,bson.E{Key: "start_date",Value: menu.Start_Date})
			UpdateObj = append(UpdateObj,bson.E{Key: "end_date",Value: menu.End_date})

			if(menu.Name != ""){
				UpdateObj = append(UpdateObj,bson.E{Key: "name",Value: menu.Name})
			}
			
			if(menu.Category != ""){
				UpdateObj = append(UpdateObj,bson.E{Key: "category",Value: menu.Category})
			}

			menu.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
			UpdateObj = append(UpdateObj,bson.E{Key : "updated_at",Value : menu.Updated_at})

			upsert := true
			opt := options.UpdateOptions{
				Upsert: &upsert,
			}

			result, err := menuCollection.UpdateOne(
				ctx,
				filter,
				bson.D{
					{Key: "$set",Value: UpdateObj},
				},
				&opt,
			)

			if(err != nil) {
				c.JSON(http.StatusInternalServerError,gin.H{"error":"Menu Update failed"})
			}
			defer cancel()
			c.JSON(http.StatusOK,result)
		}
	}
}
