package config

var Host string = "0.0.0.0"
var Port int = 7030

var KeysLimit int = 100         // max key limit of db (demo only)
var EvictionRatio float64 = 0.4 // evicts this much ratio of keys whenever eviction runs
var EvictionStrategy string = "allkeys-random"

var AOFFile string = "./sider.aof"
var RDBFile string = "./sider.rdb"

var AOFMaxSizeBytes int64 = 64 * 1024 * 1024
var SnapshotIntervalMs int64 = 30 * 1000
var AOFRewriteIntervalMs int64 = 5 * 60 * 1000

var CheckpointFile string = "./sider.checkpoint"
