package controllers

import (
	"Restaurant-Backend/database"
	"Restaurant-Backend/models"
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		//pagination stuff...
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10 //10 items per page
		}

		page, err := strconv.Atoi(c.Query("page"))

		if err != nil || page < 1 {
			page = 1 //no .of records of first page
		}

		startIndex := (page - 1) * recordPerPage //skipping limit
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		//aggregation query

		//set critera to pick data
		matchStage := bson.D{{"$match", bson.D{{}}}}

		//making groups on the basis of criteria
		groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{"_id", "null"}}},
			{Key: "total_count", Value: bson.D{{"$sum", "1"}}},
			{Key: "data", Value: bson.D{{"$push", "$$ROOT"}}},
		}}}

		//mention what need to display to UI
		projectStage := bson.D{
			{
				"$project", bson.D{
					{"_id", 0}, //id does not go
					{"total_count", 1},
					{"food_item", bson.D{
						{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
				},
			},
		}

		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing food items !"})
			log.Fatal(err)
			return
		}

		var allFoods []bson.M

		if err = result.All(ctx, &allFoods); err != nil {
			fmt.Println("something went wrong while fetching all-food in food controller!")
			// log.Fatal(err)
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		c.JSON(http.StatusOK, allFoods[0])
	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		foodId := c.Param("food_id")

		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while fetching food by Id"})
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		var food models.Food

		var menu models.Menu

		if err := c.BindJSON(&food); err != nil {
			fmt.Println("something went wrong whilte creating food!")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(food)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)

		if err != nil {
			msg := fmt.Sprintf("Menu was not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()

		var num = toFixed(*food.Price, 2)
		food.Price = &num

		result, insertErr := foodCollection.InsertOne(ctx, food)

		if insertErr != nil {
			msg := fmt.Sprintf("food itm was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)

	}
}

func round(num float64) int {
	return 	 int(num+math.Copysign(0.5,num));
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10,float64(precision));
	return float64(round(num*output))/output;
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var menu models.Menu
		var food models.Food

		foodId := c.Param("food_id")

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return;			
		}

		var updateObj primitive.D

		if food.Name!=nil{
			updateObj = append(updateObj, bson.E{"name",food.Name});
		}

		if food.Price != nil{
			updateObj = append(updateObj, bson.E{"price",food.Price});

		}

		if food.Food_image!=nil{
			updateObj = append(updateObj, bson.E{"food_image",food.Food_image});

		}	

		if food.Menu_id!=nil{	
			//check this below line later.
			err := menuCollection.FindOne(ctx,bson.E{"menu_id",food.Menu_id}).Decode(&menu)
			defer cancel();
			if err!=nil{
				msg:=fmt.Sprintf("Message : Menu was not found");
				c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
				return;
			}

			updateObj = append(updateObj,bson.E{"menu",food.Price})
		}		

		food.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at",food.Updated_at});

		upsert := true;

		filter := bson.M{"food_id":foodId}

		option := options.UpdateOptions{
			Upsert:&upsert,
		}

		result,err := foodCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set",updateObj},
			},
			&option,
		)

		if err!=nil{
			msg:=fmt.Sprint("Food item update failed");
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			return;
		}

		c.JSON(http.StatusOK,result);
	}
}
