# cd deploy
# docker-compose up &
# cd ..
export DATABASE_DSN="postgresql://localhost/gofemartdb?user=postgres&password=GofeMartPg&sslmode=disable"
go run cmd/main.go
# cd deploy
# docker-compose down
# cd ..