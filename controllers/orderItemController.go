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

type OrderItemPack struct{
	Table_id *string
	Order_items []models.OrderItemModel
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client,"orderItem");


func GetOrderItems() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		
		
		result,err := orderItemCollection.Find(context.TODO(),bson.M{});

		defer cancel();

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while listing ordered items"});
			return;
		}

		var allOrderItems []bson.M;

		if err=result.All(ctx,&allOrderItems); err!=nil{
			fmt.Println("Something went wrong in GetOrderItems()",err.Error())
			log.Fatal(err);
			return;
		}

		c.JSON(http.StatusOK,allOrderItems);
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc{
	return func (c *gin.Context){
		orderId := c.Param("order_id");
		allOrderItems,err := ItemsByOrder(orderId);


		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while listing orders from order id"});
			return;
		}

		c.JSON(http.StatusOK,allOrderItems);
	}
}

func GetOrderItem() gin.HandlerFunc{
	return func (c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);

		orderItemId := c.Param("order_item_id");

		var orderItem models.OrderItemModel;

		err := orderItemCollection.FindOne(ctx,bson.M{"order_item_id":orderItemId}).Decode(&orderItem);

		defer cancel();

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while listing order item"});
			return;
		}

		c.JSON(http.StatusOK,orderItem);
	}
}

func CreateOrderItem() gin.HandlerFunc{
	return func (c *gin.Context){
			var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
			var orderItemPack OrderItemPack

			var order models.Order;

			if err := c.BindJSON(&orderItemPack);err!=nil{
				c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});
				return;
			}
			
			order.Order_date ,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));

			orderItemsToBeInserted := []interface{}{};
			order.Table_id = orderItemPack.Table_id;
			order_id := OrderItemOrderCreator(order)

			for _,orderItem := range orderItemPack.Order_items{
				orderItem.Order_id = order_id;
				validationErr := validate.Struct(orderItem);

				if validationErr!=nil{
					c.JSON(http.StatusBadRequest,gin.H{"error":validationErr.Error()});
					return;
				}


				orderItem.ID = primitive.NewObjectID();
				orderItem.Created_at,_ =time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));

				orderItem.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));

				orderItem.Order_item_id = orderItem.ID.Hex();

				var num = toFixed(*orderItem.Unit_price,2);
				orderItem.Unit_price = &num;
				orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
			}

		  insertedOrderItems,err :=	orderItemCollection.InsertMany(ctx,orderItemsToBeInserted);

		  if err!=nil{
			fmt.Println("Something went wrong on createOrderItem()");
			log.Fatal(err);
		  }

		  defer cancel();

		  c.JSON(http.StatusOK,insertedOrderItems);

	}
}

func UpdateOrderItem() gin.HandlerFunc{
	
	return func (c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);

		var orderItem models.OrderItemModel;
		orderItemId := c.Param("order_item_id");

		filter := bson.M{"order_item_id":orderItemId};

		var updatedObj primitive.D;

		if orderItem.Unit_price!=nil{
			updatedObj = append(updatedObj, bson.E{Key: "unit_price",Value: &orderItem.Unit_price});
		}

		if orderItem.Quantity!=nil{
			updatedObj = append(updatedObj, bson.E{Key: "quantity",Value: &orderItem.Quantity});
		}

		if orderItem.Food_id!=nil{
			updatedObj = append(updatedObj, bson.E{Key: "food_id",Value: &orderItem.Food_id});
		}

		orderItem.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
		updatedObj = append(updatedObj, bson.E{"updated_at",orderItem.Updated_at});
	
		
		upsert:=true;
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result,err := orderItemCollection.UpdateOne(
			ctx,
			filter,
			bson.D{{"$set",updatedObj}},
			&opt,
		);

		if err!=nil{
			msg:= "Order item update failed -> "+err.Error();
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			return;
		}

		defer cancel();

		c.JSON(http.StatusOK,result);
	}
	
}


func ItemsByOrder(id string ) (OrderItems []primitive.M,err error){
	var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
	var orderItems []primitive.M;

	//helps to get all the order items of specific orderid
	matchStage := bson.D{{"$match",bson.D{{"order_id",id}}}}


	//looking up for food from food-collection
	lookupStage := bson.D{{"$lookup",bson.D{{"from","food"},{"localField","food_id"},{"foreignField","food_id"},{"as","food"}}}}

	unwindStage := bson.D{{"$unwind",bson.D{{"path","$food"},{"preserveNullAndEmptyArrays",true},}}}

	//looking up for order
	lookupOrderStage := bson.D{{"$lookup",bson.D{{"from","order"},{"localField","order_id"},{"foreignField","order_id"},{"as","order"},}}}

	unwindOrderStage := bson.D{{"$unwind",bson.D{{"path","$order"},{"preserveNullAndEmptyArrays",true},}}}


	//look up for table through order
	lookUpTableStage := bson.D{{"$lookup",bson.D{{"from","table"},{"localField","order.table_id"},{"foreignField","table_id"},{"as","table"},}}}
	unwindTableStage := bson.D{{"$unwind",bson.D{{"path","$table"},{"preserveNullAndEmptyArrays",true},},}}

	projectStage := bson.D{
		{
			Key: "$project",
			Value: bson.D{
				{"id",0},//setting id to zero,means the id does not move to next-stage.
				{"amount","$food.price"},
				{"total_count",1},
				{"food_name","$food.name"},
				{"food_image","$food.food_image"},
				{"table_number","$table.table_number"},
				{"table_id","$table.table_id"},
				{"order_id","$order.order_id"},
				{"price","$food.price"},
				{"quantity",1},
			},
		},
	}

	groupStage := bson.D{
		{
			Key: "$group",
				Value: bson.D{
					{"_id",
					bson.D{
						{"order_id","$order_id"},
						{"table_id","$table_id"},
						{"table_number","$table_number"},
					},
				},
						{ "payment_due",
							bson.D{{"$sum","$amount"}},
						},
						{
							"total_count",
							bson.D{{"$sum",1}},
						},
						{
							"order_items",
							bson.D{{"$push", "$$ROOT"}},
						},
					},
				},}

	
    projectStage2 := bson.D{
		{
			"$project",bson.D{
				{"id",0},
				{"payment_due",1},
				{"total_count",1},
				{"table_number","$_id.table_number"},
				{"order_items",1},
			},
		},
	}

   result,err := orderItemCollection.Aggregate(ctx,mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookUpTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2,
	});

	if err!=nil{
		fmt.Println("Something went wrong in ItemsByOrder()->",err.Error());
		return;
	}

	if err = result.All(ctx,&orderItems);err!=nil{
		fmt.Println("Something went wrong in ItemsByOrder() second time->",err.Error());
		return;
	}

	defer cancel();
	return orderItems,nil;
}