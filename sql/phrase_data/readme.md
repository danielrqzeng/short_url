# 短语说明
## 待用短语
用于阻断id生成方式误用这些短语，只要在待用库中，id生成器都会忽略不用这些短码
* 其文件对应的是available
    * available/phrase:短语，对于大于2个字符的短语，全匹配或者部分匹配都认为此短语不可用
    * available/regexp:正则匹配的表达式
* [available/phrase/njgjyycd8.txt]()为[牛津高阶英语词典第八版]()的词汇
* [available/phrase/xhzd12.txt]()为[新华字典第十二版]()的拼音
    > 其都是通过`mdict-analysis``python run_me_to_gen_txt.py`得来，原数据是其mdx文字版的数据


## 禁用短语
由于一些短语是敏感的，比如脏词，政治人物词，这些词需要被禁用
* 其文件对应的是forbid
    * forbid/phrase:短语，对于大于2个字符的短语，全匹配或者部分匹配都认为此短语不可用
    * forbid/regexp:正则匹配的表达式