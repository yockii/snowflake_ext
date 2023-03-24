# snowflake_ext
雪花算法，增加时间回拨处理

# 说明
算法中，默认参数对雪花算法的各个位数做了修改，修改点如下：

- workerId，原雪花为10，改为了5，31台集群足够了
- sequence，原雪花为12，改为了14，每ms可产生最大16383个数字，基本上也够用了

时间戳作为剩下的能剩多少给多少了，原则上除去首bit为0外，其余63位，目前分配成了3块，时间戳、worker、sequence

目前默认值 worker用了5bit，sequence用了14bit，剩余44bit就都给时间戳了，当然你也可以自定义，比如worker用8bit，sequence用12bit，那么时间戳就是43bit

# 用法
```
w, err := NewSnowflake(1)
w.NextId()

// 或者自定义参数，不使用默认值
w, err := NewSnowflakeWithConfig(1, &WorkerOption{
  BaseEpoch: uint64(1672531200000), // 2023-01-01
  WorkerIdBits: uint64(5), // 31个
  SequenceBits: uint64(14), // 16383个
  LastStamp: 0, // 可外部保存上次的时间戳，可通过w.LastStamp获取
  Sequence: 0, // 可外部保存上次的序列，可通过w.Sequence获取
})
```

# 时间回拨问题说明
算法中因为考虑了时间回拨，所以基准的时间戳实际上在这里就没有太大的反译的意义了，即通过id来计算生成的时间戳没有太大意义，时间戳仅作为确保不重复id的一道保障而存在。不要尝试从id中获取时间戳，在不回拨的情况下能够在一定程度上表示生成时间，但发生时间回拨后，在一定时期内，id无法作为准确的生成时间存在，直到系统时间逐步赶上程序时间戳为止。

# 效率
`
BenchmarkWorker_NextId-8        23138104                70.13 ns/op            0 B/op          0 allocs/op
`

具体可自行测试
