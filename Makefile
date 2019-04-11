neo4j-test:
	docker stop neo4j-test || echo ''
	docker run --rm -p 7474:7474 -p 7687:7687 -d --name neo4j-test --env=NEO4J_dbms_security_auth__enabled=false neo4j:3.5.1

test:
	go test ./...

neo4j-dev:
	docker-compose up -d elastic
	sleep 30
	docker-compose up -d neo4j kibana

clear-data:
	rm -R esdata
	rm -R neo4j-data

run:
	SESSION_SECRET="1234567890" go run cmd/server/api.go

image:
	docker build -t events-api .

image-dev:
	docker build -t events-api-dev . -f Dockerfile.dev

docker-db-run:
	docker run -d --rm -v neo4j-data:/data -p 7474:7474 -p 7687:7687 --env=NEO4J_dbms_security_auth__enabled=false --network host neo4j:3.5.1
docker-api-run:
	docker run --rm -p 3000:3000 --network host events-api-dev


swagger-ui-pull:
	docker pull swaggerapi/swagger-ui
swagger-editor-pull:
	docker pull swaggerapi/swagger-editor

swagger-ui-run:
	./swagger-ui-start.sh
swagger-editor-run:
	docker run -d -p 6020:8080 swaggerapi/swagger-editor