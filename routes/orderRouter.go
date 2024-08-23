package routes

import ("github.com/gin-gonic/gin"
	controller "Restaurant-Backend/controllers"
)

func OrderRoutes(incomingRoutes *gin.Engine){
	//get all order
	incomingRoutes.GET("/orders",controller.GetOrders())
	
	//get order by id
	incomingRoutes.GET("/orders/:order_id",controller.GetOrder())

	//create order
	incomingRoutes.POST("/orders",controller.CreateOrder())

	//update order
	incomingRoutes.PATCH("/orders/:order_id",controller.UpdateOrder())
}