FROM arm32v6/alpine:3.7
RUN mkdir -p /etc/lora-ddr
EXPOSE 8080
VOLUME /etc/lora-ddr
WORKDIR /root/
COPY lora-ddr .
ENTRYPOINT ["./lora-ddr", "--config", "/etc/lora-ddr/lora-ddr.yaml"]
