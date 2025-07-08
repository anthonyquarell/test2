.DEFAULT_GOAL := build

BINARY_NAME = svc
BUILD_PATH = cmd/build

build:
	mkdir -p $(BUILD_PATH)
	CGO_ENABLED=0 go build -o $(BUILD_PATH)/$(BINARY_NAME) cmd/main.go

clean:
	rm -rf $(BUILD_PATH)

# Установка плагинов protoc
proto-plugins:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Устанавливаем vendor-proto файлы
refresh-vendor-proto:
	rm -rf vendor-proto &&\
	git clone --depth=1 --single-branch git@github.com:mechta-market/protos.git vendor-proto
#	cd vendor-proto && rm -rf .git .idea && rm -rf e_product_v1 && find . -type f ! -name '*.proto' -delete

check-env:
	@if [ -f .env ]; then \
		echo ".env file found in project root"; \
		head -5 .env; \
	else \
		echo "WARNING: .env file not found in project root"; \
	fi

vendor-proto-dirs = vp-common
vp-%:
	mkdir -p pkg/proto
	protoc -I vendor-proto \
	--go_out pkg/proto --go_opt paths=source_relative \
		--go_opt=Mcommon/common.proto=`go list -m`/pkg/proto/common \
	--go-grpc_out pkg/proto --go-grpc_opt paths=source_relative \
	--grpc-gateway_out pkg/proto --grpc-gateway_opt paths=source_relative \
	vendor-proto/$(subst vp-,,$@)/*.proto

generate-vendor-proto-dirs: $(vendor-proto-dirs)

generate-proto-e_product_v1:
	mkdir -p pkg/proto
	protoc -I vendor-proto -I api/proto \
	--go_out pkg/proto --go_opt paths=source_relative \
		--go_opt=Mcommon/common.proto=`go list -m`/pkg/proto/common \
	--go-grpc_out pkg/proto --go-grpc_opt paths=source_relative \
	--grpc-gateway_out pkg/proto --grpc-gateway_opt paths=source_relative \
	--openapiv2_out=json_names_for_fields=false,allow_merge=true,merge_file_name=api:docs \
	api/proto/e_product/*.proto

generate-proto: generate-vendor-proto-dirs generate-proto-e_product_v1
