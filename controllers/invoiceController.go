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

type InvoiceViewFormat struct{
	Invoice_id 			string
	payment_method		string
	Order_id			string
	Payment_status		*string
	Payment_due			interface{}
	Table_number		interface{}
	Payment_due_date	time.Time
	Order_details		interface{}
}

var invoiceCollection = database.OpenCollection(database.Client,"invoice")

func GetInvoices() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		invoiceCursor, err := invoiceCollection.Find(ctx,bson.M{})
		if(err!= nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occcured while fetching the invoice data"})
			return
		}

		var allInvoices []bson.M

		if err = invoiceCursor.All(ctx,&allInvoices);err != nil{
			log.Fatal(err)
		}
		c.JSON(http.StatusOK,allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		invoiceId := c.Param("invoice_id")
		var invoice models.Invoice

		err := invoiceCollection.FindOne(ctx,bson.M{"invoice_id":invoiceId}).Decode(&invoice)
		if(err != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":" failed fetching the invoice with givrn invoice id"})
			return
		}

		var invoiceView InvoiceViewFormat
		allOrderItems, _ := ItemsByOrder(invoice.Order_id)
		invoiceView.Order_id = invoice.Order_id
		invoiceView.Payment_due_date = invoice.Payment_due_date
		invoiceView.payment_method = "null"
		if(invoice.Payment_method != nil){
			invoiceView.payment_method = *invoice.Payment_method
		}

		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Payment_status = invoice.Payment_status
		invoiceView.Payment_due = allOrderItems[0]["payment_due"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]		
		invoiceView.Order_details = allOrderItems[0]["order_items"]

		c.JSON(http.StatusOK,invoiceView)

	}
}

func CreateInvoice() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()

		var invoice models.Invoice

		if err := c.BindJSON(&invoice); err != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return
		}

		ValidateErr := validate.Struct(invoice)
		if ValidateErr != nil {
			c.JSON(http.StatusBadRequest,gin.H{"error":ValidateErr.Error()})
			return
		}

		var order models.Order
		
		err := orderCollection.FindOne(ctx,bson.M{"order_id":invoice.Order_id}).Decode(&order)

		if(err != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Order id does not exist"})
			return
		}

		status := "PENDING"
		if invoice.Payment_status == nil{
			invoice.Payment_status = &status
		}

		invoice.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		invoice.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()
		invoice.Payment_due_date, _ = time.Parse(time.RFC3339,time.Now().AddDate(0,0,1).Format(time.RFC3339))

		result,err := invoiceCollection.InsertOne(ctx,invoice)
		if(err != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Insertion error in invoice"})
			return
		}
		c.JSON(http.StatusOK,result)
	}
}

func UpdateInvoice() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
		defer cancel()
		var invoice models.Invoice
		var invoiceId = c.Param("invoice_id")
		var updateObj primitive.D

		if err := c.BindJSON(&invoice); err != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return
		}

		if invoice.Payment_method != nil{
			updateObj = append(updateObj, bson.E{Key: "payment_method",Value: invoice.Payment_method})
		}

		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{Key: "payment_status",Value: invoice.Payment_status})
		}

		invoice.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at",Value: invoice.Updated_at})
		
		

		var upsert = true
		var opt = options.UpdateOptions{
			Upsert: &upsert,
		}
		
		status := "PENDING"
		if invoice.Payment_status == nil{
			invoice.Payment_status = &status
			//updateObj = append(updateObj, bson.E{"payment_status",invoice.Payment_status})
		}
		
		filter := bson.M{"invoice_id":invoiceId}
		result,err := invoiceCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set",Value: updateObj},
			},
			&opt,
		)

		if(err != nil){
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error in updating the invoice"})
			return
		}
		c.JSON(http.StatusOK,result)
	}
}