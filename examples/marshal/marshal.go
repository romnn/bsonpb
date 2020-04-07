package main

import (
	"fmt"

	"github.com/romnnn/bsonpb"
	pb "github.com/romnnn/bsonpb/test_protos/test_objects"
)

func main() {
	marshaler := bsonpb.Marshaler{}
	omitMarshaler := bsonpb.Marshaler{Omit: bsonpb.OmitOptions{All: true}}

	someProto := &pb.Widget{RColor: []pb.Widget_Color{pb.Widget_RED}}
	fmt.Printf("----\nOriginal proto:\t\t\t %s\n", someProto)

	bson, err := marshaler.Marshal(someProto)
	if err != nil {
		fmt.Errorf("Failed to marshal with error: %s", err.Error())
	}
	fmt.Printf("Marshaled bson:\t\t\t %s\n", bson)

	bson, err = omitMarshaler.Marshal(someProto)
	if err != nil {
		fmt.Errorf("Failed to marshal with error: %s", err.Error())
	}
	fmt.Printf("Marshaled bson w/o defaults:\t %s\n", bson)
}
