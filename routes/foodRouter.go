package routes

import (
	controller "Restaurant-Backend/controllers"
	"github.com/gin-gonic/gin"
)

func FoodRoutes(incomingRoutes *gin.Engine){
	//get all foods
	incomingRoutes.GET("/foods",controller.GetFoods());

	//get all foods by id
	incomingRoutes.GET("/foods/:food_id",controller.GetFood())
	
	//create new food
	incomingRoutes.POST("/foods",controller.CreateFood())

	//update food
	incomingRoutes.PATCH("/foods/:food_id",controller.UpdateFood())

}