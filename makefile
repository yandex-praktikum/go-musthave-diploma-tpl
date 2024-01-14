genmock:
	mockgen -destination=mocks/mock_store.go -package=mocks github.com/k-morozov/go-musthave-diploma-tpl/adaptors/store Store
test:
	go test -v ./...
