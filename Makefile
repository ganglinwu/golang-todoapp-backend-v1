utest:
	go test ./server/... -v
	go test ./mongostore/... -v
	go test ./postgres_store/... -v

buildm:
	docker build -t docker-todo-backend:v1.0m -f mongo.Dockerfile .

buildp:
	docker build -t docker-todo-backend:v1.0p -f postgres.Dockerfile .

buildall:
	make buildm
	make buildp
