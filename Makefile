.PHONY: install run build reset fresh

install:
	./third_party/outline-ss-server.sh

run: install
	go run main.go start

build: install
	go build main.go -o shadowsocks

empty:
	find storage/prometheus/data -not -name '.gitignore' -delete
	docker compose restart

fresh:
	find storage/database -not -name '.gitignore' -delete
	find storage/prometheus/configs -not -name '.gitignore' -delete
	find storage/prometheus/data -not -name '.gitignore' -delete
	find storage/shadowsocks -not -name '.gitignore' -delete
	docker compose restart
