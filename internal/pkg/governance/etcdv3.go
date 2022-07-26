// gen by iyfiysi at 2021 May 19

package governance

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/spf13/viper"
	"time"
)

const (
	LeaseCloseCuzNone    = 0  //租约被关掉的原因-无
	LeaseCloseCuzInit    = -1 //租约被关掉的原因-初始化阶段出问题
	LeaseCloseCuzEtcd    = -2 //租约被关掉的原因-可能是etcd方面出问题，比如被另外一个程序revoke掉此租约了
	LeaseCloseCuzExpire  = -3 //租约被关掉的原因-过期了，比如续期不及时，然后就死翘翘了
	LeaseCloseCuzProgram = -4 //租约被关掉的原因-被程序关闭了（主动关闭）
)

//defaultEtcdV3Client 默认连接，读取配置文件得到的
func DefaultEtcdV3Client() (client *clientv3.Client, err error) {
	etcdEnable := viper.GetBool("etcd.enable")
	if !etcdEnable {
		err = fmt.Errorf("etcd not enable")
		return
	}
	etcdAddrs := viper.GetStringSlice("etcd.etcdServer")
	if len(etcdAddrs) == 0 {
		err = fmt.Errorf("etcd server not exist")
		return
	}
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: 5 * time.Second,
	})
	return
}

type EtcdType struct {
	id      string
	key     string
	leaseID clientv3.LeaseID
	cli     *clientv3.Client
	ops     []clientv3.OpOption
	ctx     context.Context
	cancel  context.CancelFunc

	//一些回调函数
	onLeaseCreate func(id clientv3.LeaseID)  //租约创建成功后回调
	onLeaseClose  func(code int, msg string) //租约失败后回调
	onKeyChange   func(key, val string)      //监控-某个key被改动了
	onKeyDelete   func(key string)           //监控-某个key被删除了
	onWatchFail   func(code int, msg string) //监控-未知错误发生了
}

//init 初始化
func (e *EtcdType) init() {
	e.leaseID = 0
	e.ops = make([]clientv3.OpOption, 0)
	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.onLeaseCreate = nil
	e.onLeaseClose = nil
	e.onKeyChange = nil
	e.onKeyDelete = nil
	e.onWatchFail = nil
}

//put 设置kv
func (e *EtcdType) put(key, val string) (err error) {
	_, err = e.cli.Put(context.TODO(), key, val, e.ops...)
	if err != nil {
		return
	}
	return
}

//del 删除某个key
func (e *EtcdType) del(key string) (err error) {
	_, err = e.cli.Delete(context.TODO(), key, e.ops...)
	if err != nil {
		return
	}
	return
}

//get 获取某个key（此处是不分基于前缀获取还是精准获取，需要设置ops值区分之)
func (e *EtcdType) get(key string) (kvs map[string]string, err error) {
	rsp, err := e.cli.Get(context.TODO(), key, e.ops...)
	if err != nil {
		return
	}
	kvs = make(map[string]string)
	for _, v := range rsp.Kvs {
		k := string(v.Key)
		v := string(v.Value)
		kvs[k] = v
	}
	return
}

// watch 监控某个key是否有变动
func (e *EtcdType) watch() {
	watchKey := e.key
	kv := clientv3.NewKV(e.cli)
	// 先GET到当前的值，并监听后续变化
	rsp, err := kv.Get(context.TODO(), watchKey, e.ops...)
	if err != nil {
		if e.onWatchFail != nil {
			e.onWatchFail(-1, err.Error())
		}
		return
	}

	nextRevision := rsp.Header.Revision + 1
	e.ops = append(e.ops, clientv3.WithRev(nextRevision))
	watcher := clientv3.NewWatcher(e.cli)

	watcherCh := watcher.Watch(e.ctx, watchKey, e.ops...)
	//// 处理kv变化事件
	//for watchResp := range watcherCh {
	//	for _, event := range watchResp.Events {
	//		switch event.Type {
	//		case mvccpb.PUT:
	//			e.onKeyChange(string(event.Kv.Key), string(event.Kv.Value))
	//			//fmt.Println(aurora.Green("WATCH").String()+"key=", watchKey, " been modify to ",
	//			//	string(event.Kv.Value), "Revision:",
	//			//	event.Kv.CreateRevision, event.Kv.ModRevision)
	//		case mvccpb.DELETE:
	//			e.onKeyDelete(string(event.Kv.Key))
	//			//fmt.Println(aurora.Green("WATCH").String()+"key=", watchKey, " been delete", "Revision:", event.Kv.ModRevision)
	//		}
	//	}
	//}

	// 处理kv变化事件
	loopDone := false
	for {
		select {
		case rsp, ok := <-watcherCh:
			//未知错误
			if !ok {
				loopDone = true
				e.onWatchFail(-1, "unknow err happen")
				break
			}
			for _, event := range rsp.Events {
				switch event.Type {
				case mvccpb.PUT:
					e.onKeyChange(string(event.Kv.Key), string(event.Kv.Value))
				case mvccpb.DELETE:
					e.onKeyDelete(string(event.Kv.Key))
				}
			}
		case <-e.ctx.Done():
			//被人工关闭了续期
			loopDone = true
			e.onWatchFail(-1, "lease been close by program")
		}
		if loopDone {
			break
		}
	}

	fmt.Println("done for grWatch")
	return
}

// lease 租约处理-一直续约知道出错或者是用户关闭
func (e *EtcdType) lease() {
	//创建一个租约
	leaseCli := clientv3.NewLease(e.cli)
	leaseRsp, err := leaseCli.Grant(context.TODO(), 5) //设置租约5s的有效期
	if err != nil {
		e.onLeaseClose(LeaseCloseCuzInit, err.Error())
		return
	}
	e.leaseID = leaseRsp.ID
	e.onLeaseCreate(e.leaseID) //通知调用方租约创建成功

	leaseCh, err := leaseCli.KeepAlive(context.TODO(), e.leaseID)
	if err != nil {
		e.onLeaseClose(LeaseCloseCuzInit, err.Error())
		return
	}

	leaseDone := false
	for {
		select {
		case ka, ok := <-leaseCh:
			//续期失败
			if !ok {
				// 举例一个场景会出现此bug：比如此程序建立了一个租约，然后另外的一个程序（比如etcdctl revoke了）取消了，会报此错
				//fmt.Println(aurora.Yellow("LEASE").String()+" lease=", leaseRsp.ID, " get not ok at", time.Now())
				leaseDone = true
				e.onLeaseClose(LeaseCloseCuzEtcd,
					"something wrong happen in etcd server,maybe key revoke by other user")
				break
			}
			//租约失效了
			if ka == nil {
				leaseDone = true
				//fmt.Println(aurora.Yellow("LEASE").String()+"lease=", leaseRsp.ID, " expire at ", time.Now())
				e.onLeaseClose(LeaseCloseCuzExpire, "lease may expire")
				break
			}
			//收到自动续期应答，告知续期成功，可以继续浪
			//fmt.Println(aurora.Yellow("LEASE").String()+"lease=", leaseRsp.ID, " success,ttl:", ka.TTL)
		case <-e.ctx.Done():
			//被人工关闭了续期
			leaseDone = true
			//fmt.Println(aurora.Yellow("LEASE").String()+"lease=", e.LeaseID, " been terminal by sys")
			e.onLeaseClose(LeaseCloseCuzProgram, "lease been close by program")
		}
		if leaseDone {
			break
		}
	}

	return
}

//get 获取某个key（此处是不分基于前缀获取还是精准获取，需要设置ops值区分之)
func (e *EtcdType) Stop() {
	e.cancel()
	return
}

//RunToKeepAlive 租约形式跑起来，返回一个租约给调用者使用，然后其回维持此租约以保证key有效
func (e *EtcdType) RunToKeepAlive(
	cli *clientv3.Client,
	key string,
	onLeaseCreate func(id clientv3.LeaseID),
	onLeaseClose func(code int, msg string)) (err error) {

	e.init()
	e.cli = cli
	e.key = key
	e.onLeaseCreate = onLeaseCreate
	e.onLeaseClose = onLeaseClose
	go e.lease()
	return
}

//RunToWatch 监控某个key
func (e *EtcdType) RunToWatch(
	cli *clientv3.Client,
	key string,
	onKeyChange func(key, val string), //监控-某个key被改动了
	onKeyDelete func(key string), //监控-某个key被删除了
	onWatchFail func(code int, msg string), //监控-未知错误发生了
) (err error) {

	e.init()
	e.cli = cli
	e.key = key
	e.onKeyChange = onKeyChange
	e.onKeyDelete = onKeyDelete
	e.onWatchFail = onWatchFail
	go e.watch()
	return
}

//RunToWatchPrefix 使用前缀，监控一堆key的变化
func (e *EtcdType) RunToWatchPrefix(
	cli *clientv3.Client,
	prefix string,
	onKeyChange func(key, val string), //监控-某个key被改动了
	onKeyDelete func(key string), //监控-某个key被删除了
	onWatchFail func(code int, msg string), //监控-未知错误发生了
) (err error) {

	e.init()
	e.cli = cli
	e.key = prefix
	e.ops = append(e.ops, clientv3.WithPrefix())
	e.onKeyChange = onKeyChange
	e.onKeyDelete = onKeyDelete
	e.onWatchFail = onWatchFail
	go e.watch()
	return
}

//Put 设置值
func (e *EtcdType) Put(cli *clientv3.Client, key, val string) (err error) {
	e.init()
	e.cli = cli
	e.key = key
	err = e.put(key, val)
	return
}

//Get 获取某个值，精准命中此值即返回，否则报错
func (e *EtcdType) Get(cli *clientv3.Client, key string) (val string, err error) {
	e.init()
	e.cli = cli
	e.key = key
	kvs, err := e.get(key)
	if err != nil {
		return
	}
	if len(kvs) != 1 {
		err = fmt.Errorf("len(val)=%d found for key=%s", len(kvs), key)
		return
	}
	for _, v := range kvs {
		val = v
	}
	return
}
