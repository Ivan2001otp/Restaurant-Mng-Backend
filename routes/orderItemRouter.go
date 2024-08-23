package routes

import ("github.com/gin-gonic/gin"
	controller "Restaurant-Backend/controllers"
)

func OrderItemRoutes(incomingRoutes *gin.Engine){
	//get all order
	incomingRoutes.GET("/orderItems",controller.GetOrderItems())
	
	//get order by id
	incomingRoutes.GET("/orderItems/:orderItem_id",controller.GetOrderItem())


	incomingRoutes.GET("orderItems-order/:order_id",controller.GetOrderItemsByOrder())
	//create order
	incomingRoutes.POST("/orderItems",controller.CreateOrderItem())

	//update order
	incomingRoutes.PATCH("/orderItems/:orderItem_id",controller.UpdateOrderItem())
}