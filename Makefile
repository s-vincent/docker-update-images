EXE:=docker-update-images
SRC:=src/*.go

all: build

build: $(SRC)
	go build -o $(EXE) $(SRC) 

clean:
	rm -f $(EXE)

