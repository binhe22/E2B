# IMPORTANT: Don't specify the FROM field here. The FROM field (with additional configuration) is injected during runtime.
# We will have a proper Devbook based image in the future.
{{ .BaseDockerfile }}

RUN apk update && apk upgrade 
RUN apk add --no-cache curl bash

WORKDIR code

# Set env vars for devbook-daemon
RUN echo RUN_CMD=bash >> /.dbkenv
# Format: RUN_ARGS=arg1 arg2 arg3
RUN echo RUN_ARGS=index.sh >> /.dbkenv
RUN echo WORKDIR=/code >> /.dbkenv
# Relative to the WORKDIR env.
RUN echo ENTRYPOINT=index.sh >> /.dbkenv

WORKDIR /
