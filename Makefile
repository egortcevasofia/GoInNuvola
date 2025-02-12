PROTOS := $(wildcard *.proto)
PB_GO_FILES := $(patsubst %.proto,%.pb.go,$(PROTOS))

all: $(PB_GO_FILES)

%.pb.go: %.proto
	C:/Users/User/protoc-29.3-win64/bin/protoc.exe --plugin=protoc-gen-go=C:/Users/User/go/bin/protoc-gen-go.exe --go_out=. $<

clean:
	rm -f $(PB_GO_FILES)