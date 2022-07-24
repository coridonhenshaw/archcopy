package rpcslave

import (
	rpc "Archcopy/RPC"
	util "Archcopy/util"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	pb "proto.local/archcopyRPC"
)

func (o *ServerStruct) Disconnect(ctx context.Context, In *pb.SessionKey) (*pb.Status, error) {
	var rc pb.Status
	var err error

	Sess := o.Housekeeping(&In.Key)

	if Sess == nil {
		return &rc, errors.New("Session failure.")
	}

	if subtle.ConstantTimeCompare(In.Signature, rpc.Sign(In.Key)) == 0 {
		return &rc, errors.New("Signature check failed.")
	}

	fmt.Printf("Ended session %v for client-id %v\n", util.Base64([]byte(In.Key)), util.Base64(Sess.ClientID))

	o.SessionTableMutex.Lock()
	delete(o.SessionTable, SessionKey(In.Key))
	o.SessionTableMutex.Unlock()

	if o.SingleSession {
		go func() {
			time.Sleep(2 * time.Second)
			o.GRPC.GracefulStop()
			fmt.Println("Terminating.")
		}()
	}

	return &rc, err
}
