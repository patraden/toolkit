# Vertica in Docker

Run Vertica locally on macOS using Colima and Docker.

## Prerequisites (macOS)
```shell
brew install colima docker qemu lima-additional-guestagents
```

## Build Vertica CE image

Since OpenText no longer publishes Vertica Community Edition (CE) images on Docker Hub, you need to build the image manually by following the official instructions in the [Vertica Containers repository](https://github.com/vertica/vertica-containers).

You can obtain the required `.rpm` installer in one of the following ways:

1. Register on the OpenText website and submit a trial request to download Vertica CE.  
2. Use your company’s internal package repository, if available.

## Run Vertica in Docker
```shell
# start the vm
colima start --arch x86_64 --cpu 6 --memory 12 --disk 100
# verify
colima list
# run vertica in docker
docker run -d \
  --name vertica \
  -p 5433:5433 \
  -p 5434:5434 \
  -e ACCEPT_EULA=Y \
  --platform linux/amd64 \
  vertica-ce:25.3.0-0
```

## Connect to Vertica
```shell
# from docker
docker exec -it vertica /opt/vertica/bin/vsql -U dbadmin -d VMart
# from host if vsql is available
vsql -U dbadmin -d VMart
```