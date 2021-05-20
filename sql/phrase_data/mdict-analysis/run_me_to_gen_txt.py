#!/bin/python
# -*- coding: UTF-8 -*-

from readmdict import MDX, MDD
import re
import unicodedata


def gen_cn():
	input_file 	= 'mdx/xhzd12.mdx'
	output_file = 'mdx/xhzd12.txt'
	mdx = MDX(input_file)
	items = mdx.items()
	datas = []
	for k,v in items:
		py= re.search(r'<py>(.*?)</py>',v,re.M|re.I)
		if py:
			w = unicode(py.group(1),"utf-8")
			pyword = unicodedata.normalize('NFKD', w).encode('ascii','ignore')
			ws = re.split('[ (),]',pyword)
			for w in ws:
				if w!='':
					datas.append(w)
	new_data = list(set(datas))
	new_data.sort(key=datas.index)
	fp = open(output_file,'w+')
	#for l in new_data:
	#	fp.write(l+'\n')
	fp.write('\n'.join(new_data))
	fp.close()

def gen_en():
	input_file 	= 'mdx/njgjyycd8.mdx'
	output_file = 'mdx/njgjyycd8.txt'
	mdx = MDX(input_file)
	items = mdx.items()
	datas = []
	for k,v in items:
		datas.append(k)
	new_data = list(set(datas))
	new_data.sort(key=datas.index)
	fp = open(output_file,'w+')
	#for l in new_data:
	#	fp.write(l+'\n')
	fp.write('\n'.join(new_data))
	fp.close()


if __name__ == '__main__':
	gen_cn()
	gen_en()

