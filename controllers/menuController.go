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

var menuCollection *mongo.Collection = database.OpenCollection(database.Client,"menu")


func inTimeSpan(start, end, check time.Time) bool{
	return start.After(time.Now()) && end.After(start)
}

func GetMenus() gin.HandlerFunc{
	
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);

		defer cancel();

		result,err := menuCollection.Find(context.TODO(),bson.M{})

		if err!=nil{
			fmt.Println("something went wrong while fetching all menus")
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while listing all menus"})
			return;
		}

		var allMenus []bson.M

		if err=result.All(ctx,&allMenus);err!=nil{
			log.Fatal(err)
		}

		c.JSON(http.StatusOK,allMenus)
	}
}

func GetMenu() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel =	context.WithTimeout(context.Background(),100*time.Second)
	
		defer cancel();


		menuId := c.Param("menu_id")

		var menu models.Menu

		err := foodCollection.FindOne(ctx,bson.M{"menu_id":menuId}).Decode(&menu)
		
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while fetching menu by Id"})
			log.Fatal(err)
			return;
		}
		c.JSON(http.StatusOK,menu)
	}
}


func CreateMenu() gin.HandlerFunc{
	return func(c *gin.Context){
		var menu models.Menu;

		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second)

		defer cancel()

		if err:= c.BindJSON(&menu);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return;
		}

		validationErr := validate.Struct(menu)
		if validationErr!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":validationErr.Error()});
			return;
		}

		menu.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		menu.Updated_at ,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		menu.ID = primitive.NewObjectID();

		menu.Menu_id = menu.ID.Hex();

		result,err:=menuCollection.InsertOne(ctx,menu);
		
		if err!=nil{
			fmt.Println("Insert error in menu");
			msg:=fmt.Sprintf("Menu Item was not created!");
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			return;
		}

		c.JSON(http.StatusOK,result);
		defer cancel();
	}
}


func UpdateMenu() gin.HandlerFunc{
	return func(c *gin.Context){

		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		defer cancel();

		var menu models.Menu;

		if err := c.BindJSON(&menu);err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return;
		}

		menuId := c.Param("menu_id");

		filter := bson.M{"menu_id":menuId};

		var updateObj primitive.D

		if menu.Start_date!=nil && menu.End_date!=nil{
			if !inTimeSpan(*menu.Start_date,*menu.End_date,time.Now()){
				msg:="kindly retype the time";
				c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
				defer cancel();
				return;
			}

			updateObj = append(updateObj, bson.E{Key:"start_date",Value:menu.Start_date});
			updateObj = append(updateObj, bson.E{Key:"end_date",Value:menu.End_date});

			if menu.Name!=""{
				updateObj=append(updateObj, bson.E{Key: "name",Value: menu.Name});
			}
			if menu.Category!=""{
				updateObj=append(updateObj, bson.E{Key: "category",Value: menu.Category});
			}

			menu.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
			
			updateObj = append(updateObj, bson.E{Key: "updated_at",Value: menu.Updated_at});

			upsert:=true

			opt := options.UpdateOptions{
				Upsert:&upsert,

			}

			result,err := menuCollection.UpdateOne(ctx,filter,bson.D{
				{Key: "$set",Value: updateObj},
			},&opt,)

			if err!=nil{
				msg := "Menu update failed!";
				c.JSON(http.StatusInternalServerError,gin.H{"error":msg})
				defer cancel();
				return;
			}

			// defer cancel();
			c.JSON(http.StatusOK,result);


		}
	}
}