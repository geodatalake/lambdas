FROM centos:7
MAINTAINER Stephen Ingram stevei@spaceflightinductries.com

ADD indexelastic .
RUN mkdir /opt/index && mv indexelastic /opt/index/indexelastic && chmod +x /opt/index/indexelastic

WORKDIR /opt/index
ENTRYPOINT [ "/opt/index/indexelastic" ]
