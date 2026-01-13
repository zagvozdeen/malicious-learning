.PHONY: build deploy

dev:
	GOEXPERIMENT=jsonv2 go run cmd/main.go

build:
	npm run build

deploy: build
	scp -r dist/* root@185.221.214.4:/var/www/ml.creavo.ru
	scp deploy/nginx.conf root@185.221.214.4:/etc/nginx/sites-available/ml.creavo.ru
	ssh root@185.221.214.4 "nginx -s reload"
