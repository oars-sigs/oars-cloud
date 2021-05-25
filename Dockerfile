FROM library/centos:7
RUN yum install -y iptables ipset 
ADD bin/oars /bin/
ENTRYPOINT ["/bin/oars"]
