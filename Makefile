#! /usr/bin/make

NAME=cacher

default: $(NAME)
$(NAME): test *.go */*.go glide.*
	go build -o $(NAME) .

cacher_cli: ./cli/*.go ./cli/*/*.go
	go build -o cacher_cli ./cli/.

test: 
	go test ./cache

bench: 
	go test ./cache -bench=.