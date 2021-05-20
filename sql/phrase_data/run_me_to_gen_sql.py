#!/bin/python
# -*- coding: UTF-8 -*-

import os
import re
import hashlib
import time


def ts():
    return int(round(time.time()))


def md5sum(s):
    m = hashlib.md5()
    m.update(s)
    return m.hexdigest()


def clean(word):
    return ''.join(filter(str.isalnum, word))

# is_subset_of_reg 判断ws是否是curr_ws_list的正则子集
def is_subset_of_reg(ws, curr_ws_list):
    wl_str = '>'.join(ws)  # 目前使用>做连接，是因为'>'是不存在于ws和curr_ws_list中的字符
    for vs in curr_ws_list:
        phrase = '.*?'.join(vs)
        phrase = '.*' + phrase + '.*'
        match = re.match(phrase, wl_str, re.I)  # 忽略大小写
        if match:
            return True, vs
    return False, []


def gen_phrase(dirs):
    # 将合法的词过滤出来
    ws_list = []
    for d in dirs:
        files = os.listdir(d)
        for f in files:
            fn = os.path.join(d, f)
            if not os.path.isfile(fn):
                continue
            for l in open(fn, 'r'):
                l = l.strip()
                if l == '':
                    continue
                ws = re.split("[ -,'.()]", l)
                ws = [clean(x) for x in ws]  # 去掉非字母和数字的字符
                ws = [x.strip() for x in ws if x.strip() != '']  # 去掉空字符
                if len(ws) <= 0:
                    continue
                # 一个字母的，忽略掉，否则太容易命中了
                # @TODO 是否两个字母的也忽略呢？
                if len(ws) == 1 and len(ws[0]) <= 1:
                    continue
                ws_list.append(ws)
    print('-' * 20)
    # 排序，数组长度短的在前，长的在后，数组长度相同的情况下，数组字符串数量短的在前
    sort_list = sorted(ws_list, key=lambda x: (len(x), len(''.join(x)), x[0]))
    # 整理正则表达式，去掉其中词语队列的子集，只保留父集（比如牛津和新华字典共有5.8w条数组词，去掉后留下来的父集只有400条）
    # 此举是为了优化正则匹配的性能，否则5.8w条组成的正则，匹配一次需要耗时30ms


    curr_ws_list = []
    for ws in sort_list:
        is_subset, parent = is_subset_of_reg(ws, curr_ws_list)
        if is_subset:
            print(','.join(ws) + ' is sub to ' + ','.join(parent))
        else:
            curr_ws_list.append(ws)
            print(','.join(ws) + ' is parent')
    # 整理出来正则数据
    datas = {}
    for ws in curr_ws_list:
        phrase = '.*?'.join(ws)
        phrase = '.*' + phrase + '.*'
        id = md5sum(phrase)
        item = {
            'id'    : id,
            'type'  : 2,
            'phrase': phrase,
        }
        datas[id] = item
    return datas


def gen_regex(dirs):
    datas = {}  # id={id,type,phrase}
    for d in dirs:
        files = os.listdir(d)
        for f in files:
            fn = os.path.join(d, f)
            if not os.path.isfile(fn):
                continue
            # 敏感词正则库
            for l in open(fn, 'r'):
                l = l.strip()
                if l == '':
                    continue
                phrase = l
                id = md5sum(phrase)
                item = {
                    'id'    : id,
                    'type'  : 2,
                    'phrase': phrase,
                }
                datas[id] = item
    return datas


def gen_sql(datas, table, output_file):
    # output_file = '../phrase_data.sql'
    str = 'INSERT IGNORE INTO {}(`phrase_type`,`phrase`,`create_ts`,`version`) values\n'.format(table)
    vals = []
    for k, v in datas.items():
        v = '({},"{}",{},0)'.format(v['type'], v['phrase'], ts())
        vals.append(v)
    str += ',\n'.join(vals) + ';'
    print(str)
    fp = open(output_file, 'w+')
    fp.write(str)
    fp.close()

if __name__ == '__main__':
    # available
    datas = []
    dp = gen_phrase(['./available/phrase'])
    dr = gen_phrase(['./available/regexp'])
    datas = dict(dp, **dr)
    gen_sql(datas, 't_available_phrase_info', '../available_phrase_data.sql')

    # forbid
    datas = []
    dp = gen_phrase(['./forbid/phrase'])
    dr = gen_phrase(['./forbid/regexp'])
    datas = dict(dp, **dr)
    gen_sql(datas, 't_forbid_phrase_info', '../forbid_phrase_data.sql')
