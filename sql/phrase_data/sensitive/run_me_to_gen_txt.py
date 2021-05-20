#!/bin/python
# -*- coding: UTF-8 -*-

from pypinyin import pinyin, lazy_pinyin, Style

def topinyin(han_words):
    han_words = unicode(han_words,'utf-8')
    py_words = lazy_pinyin(han_words)
    return py_words

def clean(word):
    return ''.join(filter(str.isalnum, word))

def gen():
    files = [
        ['corruption.txt','../forbid/regexp/corruption.txt'],
        ['livelihood.txt','../forbid/regexp/livelihood.txt'],
        ['others.txt','../forbid/regexp/others.txt'],
        ['reactionist.txt','../forbid/regexp/reactionist.txt'],
        ['sexy.txt','../forbid/regexp/sexy.txt'],
        ['terrorist.txt','../forbid/regexp/terrorist.txt'],
    ]

    for v in files:
        input_file = v[0]
        output_file = v[1]
        print('gen file={}\t\tto\t {}'.format(input_file,output_file))
        datas = []
        for l in open(input_file,'r'):
            py_words = topinyin(l)
            py_words=[s.encode('utf-8') for s in py_words]
            py_words=[clean(x) for x in py_words]
            py_words=[x.strip() for x in py_words if x.strip() != '']
            if len(py_words)==0:
                continue
            val = '.*?'.join(py_words)
            val = '.*' + val + '.*'
            datas.append(val)

        new_data = list(set(datas))
        new_data.sort(key=datas.index)
        fp = open(output_file,'w+')
        fp.write('\n'.join(new_data))
        fp.close()

if __name__ == '__main__':
	gen()

