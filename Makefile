.PHONY: dev build deploy

dev:
	GOEXPERIMENT=jsonv2 go run cmd/main.go

build:
	GOEXPERIMENT=jsonv2 go build -o malicious-learning cmd/main.go
	npm run build

deploy: build
	ssh root@185.221.214.4 "cd /var/www/ml.creavo.ru && systemctl stop malicious-learning.service && rm malicious-learning || true"
	mv malicious-learning dist/
	mv .env.prod dist/.env
	scp -r dist/* root@185.221.214.4:/var/www/ml.creavo.ru
	scp deploy/nginx.conf root@185.221.214.4:/etc/nginx/sites-available/ml.creavo.ru
	scp deploy/malicious-learning.service root@185.221.214.4:/etc/systemd/system
	ssh root@185.221.214.4 "systemctl daemon-reload && systemctl restart malicious-learning.service && nginx -s reload"
