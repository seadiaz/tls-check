FROM drone/ca-certs

ADD /tls-checker /tls-checker

ENTRYPOINT ["/tls-checker"]
