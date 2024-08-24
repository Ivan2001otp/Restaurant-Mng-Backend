package controllers

import (
	"Restaurant-Backend/database"
	"Restaurant-Backend/helper"
	"Restaurant-Backend/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)


var userCollection *mongo.Collection =  database.OpenCollection(database.Client,"user");


func GetUsers() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		
		recordPerPage,err := strconv.Atoi(c.Query("recordPerPage"))

		if err!=nil || recordPerPage<1 {
			recordPerPage = 10
		}

		page,err1 := strconv.Atoi(c.Query("page"));
		
		if err1!=nil || page<1{
			page=1;
		}

		startIndex := (page-1) * recordPerPage;

		startIndex,err3 := strconv.Atoi(c.Query("startIndex"))

		if err3!=nil{

		}

		matchStage := bson.D{{"$match",bson.D{{}}}};

		projectStage := bson.D{
			{"$project",bson.D{
				{"_id",0},
				{"total_count",0},
				{"user_items",bson.D{
					{"$slice",[]interface{}{"$data",startIndex,recordPerPage}},
				}},
			}},
		}

		result,err := userCollection.Aggregate(ctx,mongo.Pipeline{
			matchStage,
			projectStage,
		})

		defer cancel();

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while listing user items"});
			return;
		}

		var allUsers []bson.M;

		if err = result.All(ctx,&allUsers);err!=nil{
			fmt.Println("Something went wrong GetUsers() ->",err.Error());
			log.Fatal(err.Error());
			return;
		}

		c.JSON(http.StatusOK,allUsers);

	}
}

func GetUser() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);

		userId := c.Param("user_id");

		var user models.User;

	 	err := userCollection.FindOne(ctx,bson.M{"user_id":userId}).Decode(&user);

		defer cancel();

		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error while listing useritems "});
			return;
		}

		c.JSON(http.StatusOK,user);
	}
}

func Login() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		var user models.User;

		fmt.Println("exe1,");
		var foundUser models.User;

		if err:=c.BindJSON(&user);err!=nil{
			fmt.Println("Something went wrong-2->",err.Error());

			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()});
			defer cancel();
			return;
		}

		fmt.Println("exe2,");

		err := userCollection.FindOne(ctx,bson.M{"email":*user.Email}).Decode(&foundUser)
		
		fmt.Println("exe3,");
		
		defer cancel();

		fmt.Println("exe4,");

		if err!=nil{
			fmt.Println("exe-captured!");
			// fmt.Println("Something went wrong-1->",err.Err().Error());
			fmt.Println("exe-captured! 2");

			c.JSON(http.StatusInternalServerError,gin.H{"error":"User not found.Login seems to be incorrect!"});
			return;
		}

		fmt.Println("exe5,");


		passwordIsValid,msg := VerifyPassword(*user.Password,*foundUser.Password);

		if !passwordIsValid{
			fmt.Println(msg);
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			fmt.Println("exe9,");

			return;
		}

		token,refreshToken,_ := helper.GenerateAllTokens(*foundUser.Email,*foundUser.First_name,*foundUser.Last_name,foundUser.User_id)
		
		helper.UpdateAllTokens(token,refreshToken,foundUser.User_id)
		
		fmt.Println("exe16,");

		c.JSON(http.StatusOK,foundUser);
	}
}

func SignUp() gin.HandlerFunc{
	return func(c *gin.Context){

		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
		var user models.User;

		//convert json data to golang struct
		if err := c.BindJSON(&user);err!=nil{
			msg:= "something wrong in SignUp while using bindJSON->"+err.Error();
			c.JSON(http.StatusBadRequest,gin.H{"error":msg});
			defer cancel();
			return;
		}

		
		

		// validate the data
		validationErr := validate.Struct(user);

		if validationErr!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":validationErr.Error()});
			defer cancel();
			return;
		}

		//check if email exists
		count,err := userCollection.CountDocuments(ctx,bson.M{"email":user.Email});

		if err!=nil{
			fmt.Println("something went wrong in Signup in countingDocuments"+err.Error());
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while checking for the email"});
			defer cancel();
			return;
		}

		// hash password
		password := HashPassword(*user.Password);
		user.Password = &password;

		//Check if phonenum exists
		count,err = userCollection.CountDocuments(ctx,bson.M{"phone":user.Phone});

		defer cancel();

		if err!=nil{
			fmt.Println("something went wrong in Signup in countingDocuments for phone"+err.Error());
			c.JSON(http.StatusInternalServerError,gin.H{"error":"Error occured while checking for the phone"});
			return;
		}

		if count>0 {
			c.JSON(http.StatusInternalServerError,gin.H{"error":"This email or phone number already exists"});
			return;
		}

		//create some extra details for the user object-created_at,updated_at
		user.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
		user.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
		user.ID = primitive.NewObjectID();
		user.User_id = user.ID.Hex();

		//generate token & refresh token(generateAlltokens is invoked)
		token,refreshToken,_ := helper.GenerateAllTokens(*user.Email,*user.First_name,*user.Last_name,user.User_id);
	
		user.Token = &token;
		user.Refresh_token = &refreshToken;

		result,err := userCollection.InsertOne(ctx,user);
		defer cancel();

		if err!=nil{
			msg:=fmt.Sprintf("User item was not created!");
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg});
			return;
		}

		fmt.Println(result);
		c.JSON(http.StatusOK,result);

		
	}
}

func HashPassword(password string )string{
	bytes,err := bcrypt.GenerateFromPassword([]byte(password),14);

	if err!=nil{
		fmt.Println("Something went wrong in hashPasswordFunction")
		log.Panic(err);
		return "";
	}

	return string(bytes);
}

func VerifyPassword(userpassword string,providedPassword string) (bool,error){
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword),[]byte (userpassword));

	check := true
	msg:="";
	fmt.Println("exe6,");

	if err!=nil{

		msg = fmt.Sprintf("Login or password is incorrect .Something happened in VerifyPassword()");
		fmt.Println(msg);
		check = false;
		fmt.Println("exe7,");

		return check,err;
	}

	fmt.Println("exe8,");

	return check,nil;
}