package main

import (
    "log"
    "sync"
    "time"
    "github.com/google/gopacket"
    "github.com/google/gopacket/afpacket"
)

// 会话信息结构体
type FlowKey struct {
    SrcIP    string
    DstIP    string
    SrcPort  uint16
    DstPort  uint16
    Protocol uint8
}

type FlowStats struct {
    Bytes    uint64
    Packets  uint64
    LastSeen time.Time
}

// 全局变量
var (
    flowMap     = make(map[FlowKey]*FlowStats)
    flowMapLock sync.RWMutex
)

func main() {
    // 创建抓包句柄
    handle, err := afpacket.NewTPacket(
        afpacket.OptInterface("enp4s0f0"),
        afpacket.OptFrameSize(65536),
        afpacket.OptBlockSize(1<<20),
        afpacket.OptNumBlocks(128),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer handle.Close()

    // 启动多个工作协程
    var wg sync.WaitGroup
    for i := 0; i < 4; i++ {
        wg.Add(1)
        go packetProcessor(handle, &wg)
    }

    // 启动统计协程
    go statsCollector()

    wg.Wait()
} 