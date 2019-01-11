neo4j-test:
	docker stop neo4j-test || echo ''
	docker run --rm -p 7474:7474 -p 7687:7687 -d --name neo4j-test --env=NEO4J_dbms_security_auth__enabled=false neo4j:3.5.1

test:
	go test ./...

neo4j-dev:
	docker run --rm -v neo4j-data:/data -p 7474:7474 -p 7687:7687 --env=NEO4J_dbms_security_auth__enabled=false neo4j:3.5.1

run:
	SESSION_SECRET="1234567890" go run main/api.go

image:
	docker build -t events-api .
