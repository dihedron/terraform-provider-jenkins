BINARY=terraform-provider-jenkins

.DEFAULT_GOAL: $(BINARY)

$(BINARY):
	go build -o bin/$(BINARY)

test:
	go test -v

docker_test:
	go test -v
