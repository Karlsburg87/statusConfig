#!/usr/bin/env bash
#Process mangement: https://www.digitalocean.com/community/tutorials/how-to-use-bash-s-job-control-to-manage-foreground-and-background-processes

#Find if programs are installed using command built-in function
#See: http://manpages.ubuntu.com/manpages/trusty/man1/bash.1.html (search: 'Run  command  with  args')
if [[ -x "$(command -v podman)" ]]; then

    podman build -f Dockerfile -t statusconfig \
    --build-arg PROXY_TO="${PROXY_TO}" \
    --build-arg PORT="${PORT}" \
    .
    
    podman run --rm --name statusConfig -it -p 8080:"${PORT}" statusconfig

    #cleanup
    #podman image prune

elif [[ -x "$(command -v docker)" ]]; then
  
    DOCKER_BUILDKIT=1 docker build -f Dockerfile -t statusconfig \
    --progress=plain \
    --build-arg PROXY_TO="${PROXY_TO}" \
    --build-arg PORT="${PORT}" \
    .
    
    docker run --rm --name statusConfig -it -p 8080:"${PORT}" statusconfig

    #cleanup
    docker image prune
else
  echo "You need to have either Docker Desktop or Podman to run"
fi