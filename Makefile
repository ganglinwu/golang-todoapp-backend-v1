utest:
	go test ./server/... -v
	go test ./mongostore/... -v
	go test ./postgres_store/... -v
