module Archcopy

go 1.17

require (
	github.com/djherbis/times v1.5.0
	github.com/ncw/directio v1.0.5
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

require (
	github.com/DataDog/zstd v1.5.2
	github.com/dustin/go-humanize v1.0.0
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
	google.golang.org/grpc v1.47.0
	proto.local/archcopyRPC v0.0.0
)

replace proto.local/archcopyRPC => ../Protobuf
