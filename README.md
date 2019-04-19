File Server with MQ
=

How to install docker
---

 1. [Install Docker-CE (ubuntu)](https://docs.docker.com/install/linux/docker-ce/ubuntu/);
 2. [Install Docker compose](https://docs.docker.com/compose/install/);
 3. Unzip docker/docker.zip to folder(Nats 1.4.1);
 4. In this folder: `sudo docker-compose up`.

How to use proto files
---
About [gRPC](https://grpc.io/docs/). 
After install protoc and install [plugin for go](https://github.com/golang/protobuf).
Don't forget add **~/go/bin** to your PATH, just like this for example: in **~/.profile** in your home directory add this to end of file:
```bash
if [ -d "$HOME/go/bin" ] ; then
  PATH="$PATH:$HOME/go/bin"
fi
``` 
And find it: `echo $PATH`

How to generate `*.proto` to `*.go` In folder where `*.proto` input it in terminal: `protoc --go_out=. file_name.proto`
 