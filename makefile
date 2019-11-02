build:
	./download_fly.sh
	export DOCKER_BUILDKIT=0
	source ./version && docker build -t eugenmayer/concourse-pipeline-resource:"$${VERSION}" .

push:
	source ./version && docker push eugenmayer/concourse-pipeline-resource:"$${VERSION}"
