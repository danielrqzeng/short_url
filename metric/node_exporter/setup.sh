# gen by iyfiysi at 2021 Jun 16

# setup node_exporter to monitor machine


nohup ./node_exporter --web.listen-address=":9100" >/dev/null 2>&1 &