FROM centos:7
MAINTAINER Steve Ingram stevei@spaceflightinductries.com

ADD processgeo .
RUN mkdir /opt/processgeo && mv processgeo /opt/processgeo/processgeo && chmod +x /opt/processgeo/processgeo

WORKDIR /opt/processgeo
ENTRYPOINT [ "/opt/processgeo/processgeo" ]
