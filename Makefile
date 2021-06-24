build:
	docker build -t build_env -f Dockerfile-build .
	docker build -t megamon:latest -f Dockerfile .

clean:
	docker image rm build_env
