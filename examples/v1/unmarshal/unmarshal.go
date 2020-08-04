package main

import (
	"fmt"

	"github.com/romnnn/bsonpb"
	pb "github.com/romnnn/bsonpb/test_protos/test_objects"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	unmarshaler := bsonpb.Unmarshaler{}

	widgetBson := bson.D{{"rColor", bson.A{"RED"}}}
	fmt.Printf("----\nOriginal bson:\t %s\n", widgetBson)

	// Marshal bson to bytes
	rawBson, err := bson.Marshal(widgetBson)
	if err != nil {
		fmt.Printf("marshaling bson to bytes failed: %s", err.Error())
	}

	var result pb.Widget
	err = unmarshaler.Unmarshal(rawBson, &result)
	if err != nil {
		fmt.Printf("unmarshaling failed: %s\n", err.Error())
	}
	fmt.Printf("Unmarshaled proto:\t %s\n", result)
}
