FROM krakend:2

COPY . /etc/krakend/
ENV FC_ENABLE=1
ENV FC_PARTIALS=/etc/krakend/partials

EXPOSE 8000

CMD ["run", "-c", "/etc/krakend/krakend.tmpl"]
