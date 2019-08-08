build:
	./download_fly.sh
	export DOCKER_BUILDKIT=0
	docker build -t eugenmayer/concourse-pipeline-resource .

push:
	docker push eugenmayer/concourse-pipeline-resource