package rpcslave

import (
	rpc "Archcopy/RPC"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"log"
	"time"

	util "Archcopy/util"

	pb "proto.local/archcopyRPC"
)

func (s *ServerStruct) Connect(ctx context.Context, in *pb.ConnectInquire) (*pb.ConnectResponse, error) {
	s.Housekeeping(nil)

	s.SessionTableMutex.Lock()
	defer s.SessionTableMutex.Unlock()

	if len(in.ClientID) != 32 {
		log.Print("Wrong ClientID length.")
		return &pb.ConnectResponse{}, nil
	}

	var CTKey = string(in.ClientID)

	PreSession, PSValid := s.ClientTable[CTKey]
	if !PSValid {
		if in.Phase == 0 {

			Candidate := in.OfferSignature
			Expected := rpc.Sign(in.OfferNonce)

			if subtle.ConstantTimeCompare(Candidate, Expected) == 0 {
				log.Print("Client used wrong PSK during inquire phase.")
				return &pb.ConnectResponse{}, nil
			}

			Retval := new(pb.ConnectResponse)
			Challenge := make([]byte, 64)
			rand.Read(Challenge)
			Retval.Challenge = Challenge
			Retval.NonceSignature = rpc.Sign(in.OfferNonce)
			Retval.ChallengeSignature = rpc.Sign(Challenge)

			Retval.Phase = 0

			s.ClientTable[CTKey] = &ClientTableStruct{Phase: 1,
				Challenge: Challenge,
				Expires:   time.Now().Add(2 * time.Minute)}

			return Retval, nil
		}
	} else {
		if PreSession.Expires.Before(time.Now()) {
			log.Print("Handshake expired.")
			delete(s.ClientTable, CTKey)
			return &pb.ConnectResponse{}, nil
		}
		if PreSession.Phase != int(in.Phase) || PreSession.Phase != 1 {
			log.Print("Handshake sync error or invalid handshake")
			return &pb.ConnectResponse{}, nil
		}

		Candidate := in.ChallengeResponse
		Expected := rpc.Sign(PreSession.Challenge)

		delete(s.ClientTable, CTKey)

		if subtle.ConstantTimeCompare(Candidate, Expected) == 0 {
			log.Print("Client used wrong PSK during challenge phase.")
			return &pb.ConnectResponse{}, nil
		}
		FinalKey := make([]byte, 32)
		rand.Read(FinalKey)
		Session := SessionStruct{ClientID: in.ClientID, Expires: time.Now().Add(SessionLife)}
		s.SessionTable[SessionKey(FinalKey)] = &Session

		Retval := new(pb.ConnectResponse)
		Retval.Phase = 1
		Retval.Key = FinalKey
		fmt.Printf("Client-id %v authorized for session %v\n", util.Base64([]byte(CTKey)), util.Base64(FinalKey))
		return Retval, nil
	}
	// Should be unreachable
	return nil, nil
}
