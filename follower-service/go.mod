module follower-service

go 1.22

require (
	github.com/IvanNovakovic/SOA_Proj/protos v0.0.0
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/gorilla/mux v1.8.1
	github.com/neo4j/neo4j-go-driver/v5 v5.0.0
	google.golang.org/grpc v1.70.0
)

require (
	golang.org/x/net v0.32.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241202173237-19429a94021a // indirect
	google.golang.org/protobuf v1.36.0 // indirect
)

replace github.com/IvanNovakovic/SOA_Proj/protos => ../protos
