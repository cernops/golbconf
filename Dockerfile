FROM cern/c8-base
WORKDIR /root/
LABEL maintainer="LB-experts <lb-experts@cern.ch>"
RUN  dnf -y  install "dnf-command(config-manager)" && \
     dnf config-manager --add-repo  http://linuxsoft.cern.ch/internal/repos/lb8-stable/x86_64/os  && \
     yum install -y lbconf
#CMD ["sleep","3600"]
CMD ["/usr/bin/lbconf", "-configdir ", "/root", "-partition", "golang"]
