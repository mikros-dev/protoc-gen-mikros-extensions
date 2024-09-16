PLUGIN_NAME := domain_improve
SRC := main.go

PROTO := $(PLUGIN_NAME).proto
TEMPLATES := $(PLUGIN_NAME).tmpl

OUTPUT := $(PLUGIN_NAME).so
BUILD_CMD := go build -buildmode=plugin -o $(OUTPUT) $(SRC)
GEN_GO := $(PROTO:.proto=.pb.go)

all: $(OUTPUT)

proto: $(PROTO)
	protoc --go_out=. --go_opt=paths=source_relative $(PROTO)

$(OUTPUT): $(SRC) $(TEMPLATES)
	$(BUILD_CMD)

clean:
	rm -f $(OUTPUT) $(GEN_GO)

.PHONY: all clean
