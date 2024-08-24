package helper

import (
	"Restaurant-Backend/database"
	"context"
	"fmt"
	"log"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct{
	Email string
	First_name string
	Last_name string
	Uid string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client,"user");

var SECRET_KEY string = "A7d3Rz9pQ2wVb8Xs";

func GenerateAllTokens(email string,first_name string,last_name string, uid string)(signedToken string ,refreshToken string ,err error){
	claims:= &SignedDetails{
		Email: email,
		First_name: first_name,
		Last_name: last_name,
		Uid: uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt:time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt:time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token,err1 := jwt.NewWithClaims(jwt.SigningMethodHS256,claims).SignedString([]byte(SECRET_KEY))
	
	if err1!=nil{
		fmt.Println("exe13,");

		fmt.Println("something went wrong while generating signedToken");
		log.Panic(err1);
		return;
	}
	
	refreshToken,err2 := jwt.NewWithClaims(jwt.SigningMethodHS256,refreshClaims).SignedString([]byte(SECRET_KEY));

	fmt.Println("exe10,");

	if err2!=nil{
		fmt.Println("exe11,");

		log.Panic(err2);
		return;
	}

	fmt.Println("exe12,");

	return token,refreshToken,nil;
}


func UpdateAllTokens(signedToken string,signedRefreshToken string,userId string){
	var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second);
	var updateObj primitive.D;

	updateObj = append(updateObj,bson.E{"token",signedToken});
	updateObj = append(updateObj,bson.E{"refresh_token",signedRefreshToken});

	Updated_at,_ :=  time.Parse(time.RFC3339,time.Now().Format(time.RFC3339));
	updateObj = append(updateObj, bson.E{"updated_at",Updated_at})

	upsert := true;
	filter := bson.M{"user_id":userId};

	option := options.UpdateOptions{
		Upsert:&upsert,
	}

	_,err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set",updateObj},
		},
		&option,
	)

	defer cancel();

	if err!=nil{
		fmt.Println("exe14,");

		fmt.Println("something wrong in UpdateAllTokens()");
		log.Panic(err.Error());
		return;
	}

	fmt.Println("exe15,");

	return;
}

func ValidateToken(signedToken string)(claims *SignedDetails,msg string){
	token,err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY),nil;
		},
	)

	if err!=nil{
		fmt.Println("jwt.ParseWithClaims is thrown error in ValidateToken() !");
		return;
	}

	claims,ok := token.Claims.(*SignedDetails);

	if !ok{
		msg = fmt.Sprintf("The token is invalid");
		fmt.Println(msg);
		return;
	}

	if claims.ExpiresAt<time.Now().Local().Unix(){
		msg = "Token is expired";
		return;
	}

	return claims,msg;

}