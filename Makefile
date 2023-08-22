.PHONY: install run build reset fresh

install:
	./third_party/outline-ss-server.sh

run: install
	go run main.go start

build: install
	go build main.go -o shadowsocks

empty:
	find storage/prometheus/data -mindepth 1 -not -name '.gitignore' -exec rm -rf {} \;
	docker compose restart

fresh:
	find storage/database -mindepth 1 -not -name '.gitignore' -exec rm -rf {} \;
	find storage/prometheus/configs -mindepth 1 -not -name '.gitignore' -exec rm -rf {} \;
	find storage/prometheus/data -mindepth 1 -not -name '.gitignore' -exec rm -rf {} \;
	find storage/shadowsocks -mindepth 1 -not -name '.gitignore' -exec rm -rf {} \;
	docker compose restart
