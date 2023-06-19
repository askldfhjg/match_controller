FROM alpine
ADD match_controller /match_controller
ENTRYPOINT [ "/match_controller" ]
