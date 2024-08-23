package routes

import ("github.com/gin-gonic/gin"
	controller "Restaurant-Backend/controllers"
)

func MenuRoutes(incomingRoutes *gin.Engine){
	//get all invoices
	incomingRoutes.GET("/menus",controller.GetMenus())
	
	//get invoice by id
	incomingRoutes.GET("/menus/:menu_id",controller.GetMenu())

	//create invoice
	incomingRoutes.POST("/menus",controller.CreateMenu())

	//update invoice
	incomingRoutes.PATCH("/menus/:menu_id",controller.UpdateMenu())
}