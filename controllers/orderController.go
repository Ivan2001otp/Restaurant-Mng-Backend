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

var orderCollection *mongo.Collection = database.OpenCollection(database.Client,"order");

func GetOrders() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);

		result,err := orderCollection.Find(context.TODO(),bson.M{})
		defer cancel();


		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while listing order items"});
			return;
		}

		var allOrders []bson.M;

		if err = result.All(ctx,&allOrders);err!=nil{
			fmt.Println("Something went wrong while fetching all orders in orderController.");

			log.Fatal(err)
			return;
		}

		c.JSON(http.StatusOK,allOrders[0]);

	}
}

func GetOrder() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		
		var orderId = c.Param("order_id");

		var order models.Order;

		err := orderCollection.FindOne(ctx,bson.M{"order_id":orderId}).Decode(&order);

		defer cancel();

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while getting order by id"});
			return;
		}

		c.JSON(http.StatusOK,order);
	}
}


func CreateOrder() gin.HandlerFunc{
	return func(c *gin.Context){
		var table models.Table;
		var order models.Order;
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		defer cancel();

		if err:= c.BindJSON(&order);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});
			return;
		}

		validationErr := validate.Struct(order);

		if validationErr!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":validationErr.Error()});
			return;
		}


		if(order.Table_id!=nil){
			err:=tableCollection.FindOne(ctx,bson.M{"table_id":order.Table_id}).Decode(&table);
			if err!=nil{
				msg:=fmt.Sprintf("Message:Table was not found!");
				c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
				return;
			}
		}

		order.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
		order.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));

		order.ID = primitive.NewObjectID();
		order.Order_id = order.ID.Hex();

		result,err := orderCollection.InsertOne(ctx,order);
	
		if err!=nil{
			msg:=fmt.Sprintf("Order item was not created!");
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			return;
		}

		c.JSON(http.StatusOK,result);
	}
}

func UpdateOrder() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		defer cancel();


		var table models.Table;
		var order models.Order;

		var updateObj primitive.D;

		orderId:=c.Param("order_id");

		if err := c.BindJSON(&order);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});
			return;
		}


		if order.Table_id!=nil{
			err := orderCollection.FindOne(ctx,bson.E{"table_id",order.Table_id}).Decode(&table)

			if err!=nil{
				msg := fmt.Sprintf("Message:Menu was not found");
				c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
				return;
			}

			updateObj = append(updateObj, bson.E{"menu",order.Table_id})


		}

		order.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
		updateObj = append(updateObj, bson.E{"updated_at",order.Updated_at});

		upsert := true;

		filter := bson.M{"order_id":orderId};

		option := options.UpdateOptions{
			Upsert:&upsert,
		};

		result ,err := orderCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set",updateObj},
			},
			&option,
		)


		if err!=nil{
			msg:=fmt.Sprintf("Order item update failed");
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			return;
		}

		c.JSON(http.StatusOK,result);
	}
}

func OrderItemOrderCreator(order models.Order) string{
	order.Updated_at ,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
	order.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
	var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);

	defer cancel();
	
	order.ID = primitive.NewObjectID();
	_,err := orderCollection.InsertOne(ctx,order);

	if err!=nil{
		fmt.Println("Something went wrong in OrderItemOrderCreator function!");
		return "";
	}

	return order.Order_id;
}