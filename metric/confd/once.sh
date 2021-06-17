# gen by iyfiysi at 2021 Jun 16

# confd once for iyfiysi.com/short_url
# you may want to change ectd server addr if `http://127.0.0.1:2379` not the right addr

# once模式:从etcd拉取配置，生成后退出，只执行一次

confd -confdir . -config-file "./conf.d/etcd.toml" -log-level=debug -onetime -backend etcdv3 -node http://127.0.0.1:2379