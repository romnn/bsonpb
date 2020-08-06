package main

import (
	pb "github.com/romnnn/bsonpb/internal/testprotos/v1/test_objects"
	"google.golang.org/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"github.com/romnnn/bsonpb/v1"
)

func main() {
	someProto := &pb.Widget{RColor: []pb.Widget_Color{pb.Widget_RED}}
	log.Infof("Original message proto: %v", someProto)

	// Marshal
	m := &bsonpb.Marshaler{}
	marshaled, err := m.Marshal(someProto)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Message proto marshaled: %v", marshaled)

	// Unmarshal back
	um := &bsonpb.Unmarshaler{}
	var someProto2 pb.Widget
	if err := um.Unmarshal(marshaled, &someProto2); err != nil {
		log.Fatal(err)
	}
	log.Infof("Message proto unmarshaled: %v", someProto2)

	// Compare
	if !proto.Equal(someProto, &someProto2) {
		log.Fatal("Protos are not equal")
	}
	log.Info("Protos are equal")
}