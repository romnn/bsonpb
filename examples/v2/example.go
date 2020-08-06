package main

import (
	pb "github.com/romnnn/bsonpb/internal/testprotos/v2/proto3_proto"
	// "google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	// "go.mongodb.org/mongo-driver/bson"
	log "github.com/sirupsen/logrus"
	"github.com/romnnn/bsonpb/v2"
)

func main() {
	
	opts := bsonpb.MarshalOptions{}
	someProto := &pb.Message{Name: "Test", Hilarity: pb.Message_SLAPSTICK}
	log.Infof("Original proto: %s", someProto)

	marshaled, err := opts.Marshal(someProto)
	if err != nil {
		panic(err)
	}
	log.Infof("%s", marshaled)

	// Unmarshal
	target := wrapperspb.BoolValue{}
	err = bsonpb.Unmarshal(true, &target)
	log.Infof("Unmarshal: %v (%v)", target, err)
	/*
	if _, m, err := bson.MarshalValue(true); err == nil {
		err := bsonpb.Unmarshal(m, &target)
		log.Infof("Unmarshal: %v (%v)", target, err)
	}
	*/
	

	// t := timestamppb.Timestamp{}
	// log.Info(t)
	// log.Info(timestamppb.Now())

	/*
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
	*/
}