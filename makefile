include app.env

cert:
	rm -rf certificates
	mkdir certificates
	#openssl genpkey -algorithm RSA -out certificates/private.key -aes256
	#openssl req -x509 -key $(CERTIFICATE_PATH)/$(CERTIFICATE_KEY) -out $(CERTIFICATE_PATH)/$(CERTIFICATE_FILE) -days 365
	#openssl x509 -in $(CERTIFICATE_PATH)/$(CERTIFICATE_FILE) -text -noout
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout $(CERTIFICATE_PATH)/$(CERTIFICATE_KEY) -out $(CERTIFICATE_PATH)/$(CERTIFICATE_FILE) \
    -subj "/C=US/ST=State/L=City/O=Organization/OU=Department/CN=localhost"



migrateup:
	migrate -path db/migrations -database "$(CONNECTION_STRING)" -verbose up $(times)

migratedown:
	migrate -path db/migrations -database "$(CONNECTION_STRING)" -verbose down $(times)  


migrateversion:
	migrate -path db/migrations -database "$(CONNECTION_STRING)" version


new:
	migrate create -ext sql -dir db/migrations -seq $(name)

docs:
	dbdocs build doc/db.dbml

schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

mock:
	mockgen -package mockdb -destination db/mock/store.go $(PROJECT_NAME)/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go $(PROJECT_NAME)/worker TaskDistributor

gen:
	sqlc generate

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=FileConversionApi \
	proto/*.proto
	statik -src=./doc/swagger -dest=./doc

test:
	go test -v -cover -short ./...

build:
	rm -rf $(BUILD_DIR)
	mkdir $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(EXE)

clean:
	rm -rf $(BUILD_DIR)

.PHONY: network postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration db_docs db_schema sqlc test server mock proto evans redis