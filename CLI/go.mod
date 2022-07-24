module ArchcopyCLI

go 1.17

require (
	github.com/DataDog/zstd v1.5.2 // indirect
	github.com/djherbis/times v1.5.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/integrii/flaggy v1.5.2
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	proto.local/archcopyRPC v0.0.0 // indirect
)

require Archcopy v0.0.0

require github.com/ncw/directio v1.0.5 // indirect

replace Archcopy => ../Core

replace proto.local/archcopyRPC => ../Protobuf
