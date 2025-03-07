package server

// 由于cpu的pprof是持续一段时间的，需要做一个互斥
// 如果发现有正在采集的addr，则进来一个请求直接拒绝
