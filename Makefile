#! /usr/bin/make

NAME=cacher

default: $(NAME)
$(NAME): *.go */*.go glide.*
	go build -o $(NAME) .

test: 
	go test ./cache

bench: 
	go test ./cache -bench=.