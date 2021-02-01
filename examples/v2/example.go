package main

import (
	pb "github.com/romnn/bsonpb/internal/testprotos/v2/proto3_proto"
	"github.com/romnn/bsonpb/v2"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

func main() {
	someProto := &pb.Message{Name: "Test", Hilarity: pb.Message_SLAPSTICK}
	log.Infof("Original message proto: %v", someProto)

	// Marshal
	opts := bsonpb.MarshalOptions{}
	marshaled, err := opts.Marshal(someProto)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Message proto marshaled: %v", marshaled)

	// Unmarshal back
	var someProto2 pb.Message
	if err := bsonpb.Unmarshal(marshaled, &someProto2); err != nil {
		log.Fatal(err)
	}
	log.Infof("Message proto unmarshaled: %v", someProto2)

	// Compare
	if !proto.Equal(someProto, &someProto2) {
		log.Fatal("Protos are not equal")
	}
	log.Info("Protos are equal")
}
