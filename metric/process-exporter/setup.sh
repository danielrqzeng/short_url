# gen by iyfiysi at 2021 Jun 16

# setup process-exporter to monitor app short_url's process

nohup ./process-exporter -config.path process.yml -web.listen-address=":9256" >/dev/null 2>&1 &