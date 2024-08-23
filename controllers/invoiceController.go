package controllers

import (
	"Restaurant-Backend/database"
	"Restaurant-Backend/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InvoiceViewFormat struct{
	Invoic_id				string
	Payment_method			string
	Order_id				string
	Payment_status			*string
	Payment_due				interface{}
	Table_number			interface{}
	Payment_due_date		time.Time
	Order_details			interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client,"invoice");


func GetInvoices() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		
		result ,err := invoiceCollection.Find(context.TODO(),bson.M{});
		
		defer cancel();

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while listing invoice items"});
			return;
		}

		var allInvoices []bson.M;

		if err=result.All(ctx,&allInvoices);err!=nil{
			var msg string = "Something went wrong in GetInvoices() "+err.Error();
			log.Fatal(msg);
			return;
		}

		c.JSON(http.StatusOK,allInvoices);
	}
}

func GetInvoice() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		defer cancel();

		invoiceId := c.Param("invoice_id");

		var invoice models.Invoice;

		err := invoiceCollection.FindOne(ctx,bson.M{"invoice_id":invoiceId}).Decode(&invoice);

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while listing invoice item"});
			return;
		}

		var invoiceView InvoiceViewFormat

		allOrderItems,err := ItemsByOrder(invoice.Order_id);

		invoiceView.Order_id = invoice.Order_id;
		invoiceView.Payment_due_date = invoice.Payment_due_date;
		invoiceView.Payment_method = "null";

		if invoice.Payment_method!=nil{
			invoiceView.Payment_method = *invoice.Payment_method;
		}
	
		invoiceView.Invoic_id= invoice.Invoice_id;
		invoiceView.Payment_status = invoice.Payment_status;
		invoiceView.Payment_due = allOrderItems[0]["payment_due"];
		invoiceView.Table_number = allOrderItems[0]["table_number"];
		invoiceView.Order_details = allOrderItems[0]["order_items"];
	
		c.JSON(http.StatusOK,invoiceView);
	}	

}


func CreateInvoice() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		defer cancel();

		var invoice models.Invoice;

		if err:= c.BindJSON(&invoice);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});
			return;
		}

		var order models.Order;

		err := orderCollection.FindOne(ctx,
			bson.M{"order_id":invoice.Order_id,}).Decode(&order);

		if err!=nil{
			msg := fmt.Sprintf("Order was not found!");
			msg = msg+" "+err.Error();
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			return;
		}
		
		status := "PENDING";
		if invoice.Payment_status!=nil{
			invoice.Payment_status = &status;
		}

		invoice.Payment_due_date,_ = time.Parse(time.RFC3339,time.Now().AddDate(0,0,1).Format(time.RFC3339));

		invoice.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
		invoice.Updated_at ,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
		
		invoice.ID = primitive.NewObjectID();
		invoice.Invoice_id = invoice.ID.Hex();

		validationErr := validate.Struct(invoice);

		if validationErr!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":validationErr});
			return;
		}

		result,err := invoiceCollection.InsertOne(ctx,invoice);

		if err!=nil{
			msg:=fmt.Sprintf("Invoice item was not creted ->")+err.Error();
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			return;
		}

		c.JSON(http.StatusOK,result);
	}
}

func UpdateInvoice() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		defer cancel();

		var invoice models.Invoice;

		invoiceId := c.Param("invoice_id");

		if err := c.BindJSON(&invoice);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});
			return;
		}

		filter := bson.M{"invoice_id":invoiceId};
		var updateObj primitive.D;

		if invoice.Payment_method!=nil{
			updateObj = append(updateObj, bson.E{"payment_method",invoice.Payment_method});
			
		}
		if invoice.Payment_status!=nil{
			updateObj = append(updateObj, bson.E{"payment_status",invoice.Payment_status});
		}

		invoice.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
		updateObj = append(updateObj, bson.E{
			Key:   "updated_at",
			Value: invoice.Updated_at,
		})

		upsert := true;
		option := options.UpdateOptions{
			Upsert:&upsert,
		}

		status:="PENDING";

		if invoice.Payment_status==nil{
			invoice.Payment_status = &status;
		}

		result,err := invoiceCollection.UpdateOne(ctx,
			filter,
			bson.D{
				{"$set",updateObj},
			},
			&option,
		)

		if err!=nil{
			msg:= fmt.Sprintf("Ivoice item update failed!");
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			return;
		}

		c.JSON(http.StatusOK,result);
	}
}