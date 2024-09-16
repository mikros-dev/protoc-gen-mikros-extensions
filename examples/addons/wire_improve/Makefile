PLUGIN_NAME := wire_improve
SRC := main.go
TEMPLATES := $(PLUGIN_NAME).tmpl
OUTPUT := $(PLUGIN_NAME).so
BUILD_CMD := go build -buildmode=plugin -o $(OUTPUT) $(SRC)

all: $(OUTPUT)

$(OUTPUT): $(SRC) $(TEMPLATES)
	$(BUILD_CMD)

clean:
	rm -f $(OUTPUT)

.PHONY: all clean
