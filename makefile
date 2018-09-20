build:
	./download_fly.sh
	docker build -t eugenmayer/concourse-pipeline-resource .

push:
	docker push eugenmayer/concourse-pipeline-resource