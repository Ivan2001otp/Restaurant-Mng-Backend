package routes

import ("github.com/gin-gonic/gin"
	controller "Restaurant-Backend/controllers"
)

func InvoiceRoutes(incomingRoutes *gin.Engine){
	//get all invoices
	incomingRoutes.GET("/invoices",controller.GetInvoices())
	
	//get invoice by id
	incomingRoutes.GET("/invoices/:invoice_id",controller.GetInvoice())

	//create invoice
	incomingRoutes.POST("/invoices",controller.CreateInvoice())

	//update invoice
	incomingRoutes.PATCH("/invoices/:invoice_id",controller.UpdateInvoice())
}