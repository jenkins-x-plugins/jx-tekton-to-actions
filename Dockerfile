FROM centos:7

RUN yum install -y git

ENTRYPOINT ["jx-tekton-to-actions"]

COPY ./build/linux/jx-tekton-to-actions /usr/bin/jx-tekton-to-actions