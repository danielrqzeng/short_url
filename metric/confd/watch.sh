# gen by iyfiysi at 2021 Jun 16

# confd watch for iyfiysi.com/short_url
# you may want to change ectd server addr if `http://127.0.0.1:2379` not the right addr

# watch模式:后台进程永远执行，并且监控key是否变更，若是变更则重新生成配置


nohup confd -confdir . -config-file "./conf.d/etcd.toml" -log-level=debug -watch -backend etcdv3 -node http://127.0.0.1:2379 &