package routes

import ("github.com/gin-gonic/gin"
	controller "Restaurant-Backend/controllers"
)

func TableRoutes(incomingRoutes *gin.Engine){
	//get all table
	incomingRoutes.GET("/tables",controller.GetTables())
	
	//get table by id
	incomingRoutes.GET("/tables/:table_id",controller.GetTable())

	//create table
	incomingRoutes.POST("/tables",controller.CreateTable())

	//update table
	incomingRoutes.PATCH("/tables/:table_id",controller.UpdateTable())
}