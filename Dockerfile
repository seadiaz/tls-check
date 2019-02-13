FROM drone/ca-certs

ADD /tls-check /tls-check

ENTRYPOINT ["/tls-check"]
